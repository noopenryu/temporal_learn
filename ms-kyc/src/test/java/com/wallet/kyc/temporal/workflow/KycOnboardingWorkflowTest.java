package com.wallet.kyc.temporal.workflow;

import com.wallet.kyc.domain.DocumentType;
import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.domain.WalletTier;
import com.wallet.kyc.temporal.activity.*;
import com.wallet.kyc.temporal.model.*;
import io.temporal.client.WorkflowClient;
import io.temporal.client.WorkflowOptions;
import io.temporal.testing.TestWorkflowEnvironment;
import io.temporal.worker.Worker;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import java.time.LocalDate;
import java.util.concurrent.CompletableFuture;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class KycOnboardingWorkflowTest {

    private static final String TASK_QUEUE = "TEST_KYC_TASK_QUEUE";

    private TestWorkflowEnvironment testEnv;
    private Worker worker;

    private DocumentVerificationActivity documentActivity;
    private BiometricVerificationActivity biometricActivity;
    private SanctionsScreeningActivity sanctionsActivity;
    private RiskAssessmentActivity riskActivity;
    private KycPersistenceActivity persistenceActivity;
    private NotificationActivity notificationActivity;

    @BeforeEach
    void setUp() {
        testEnv = TestWorkflowEnvironment.newInstance();
        worker = testEnv.newWorker(TASK_QUEUE);
        worker.registerWorkflowImplementationTypes(KycOnboardingWorkflowImpl.class);

        documentActivity = Mockito.mock(DocumentVerificationActivity.class);
        biometricActivity = Mockito.mock(BiometricVerificationActivity.class);
        sanctionsActivity = Mockito.mock(SanctionsScreeningActivity.class);
        riskActivity = Mockito.mock(RiskAssessmentActivity.class);
        persistenceActivity = Mockito.mock(KycPersistenceActivity.class);
        notificationActivity = Mockito.mock(NotificationActivity.class);

        worker.registerActivitiesImplementations(
                documentActivity,
                biometricActivity,
                sanctionsActivity,
                riskActivity,
                persistenceActivity,
                notificationActivity
        );

        testEnv.start();
    }

    @AfterEach
    void tearDown() {
        testEnv.close();
    }

    private KycWorkflowRequest createStandardRequest(String applicationId, String customerId) {
        return new KycWorkflowRequest(
                applicationId,
                customerId,
                "Alice Smith",
                LocalDate.of(1995, 6, 20),
                "SGP",
                "+6591234567",
                "alice@example.com",
                DocumentType.PASSPORT,
                "E12345678",
                "https://cdn.wallet.local/docs/alice-front.jpg",
                null,
                "https://cdn.wallet.local/selfies/alice.jpg"
        );
    }

    @Test
    @DisplayName("Should auto-approve applicant when documents, biometrics and sanctions are clear and risk is low")
    void testAutoApprovalFlow() {
        KycWorkflowRequest request = createStandardRequest("APP-001", "CUST-001");

        when(documentActivity.verifyDocument(any()))
                .thenReturn(new DocumentVerificationResult(true, null, 98.0, "E12345678"));
        when(biometricActivity.verifyBiometrics(any()))
                .thenReturn(new BiometricVerificationResult(true, 92.0, true, "Face match high"));
        when(sanctionsActivity.screenSanctions(any()))
                .thenReturn(new SanctionsScreeningResult(true, false, false, "None"));
        when(riskActivity.assessRisk(any(), any(), any()))
                .thenReturn(new RiskAssessmentResult(15, true, false, "LOW", "All clear"));

        WorkflowOptions options = WorkflowOptions.newBuilder()
                .setTaskQueue(TASK_QUEUE)
                .setWorkflowId("wf-test-auto-approve")
                .build();

        KycOnboardingWorkflow workflow = testEnv.getWorkflowClient().newWorkflowStub(KycOnboardingWorkflow.class, options);
        KycWorkflowResult result = workflow.executeKyc(request);

        assertNotNull(result);
        assertEquals(KycStatus.APPROVED, result.getFinalStatus());
        assertEquals(WalletTier.TIER_2_PREMIUM, result.getWalletTier());
        assertEquals(15, result.getRiskScore());
        assertFalse(result.isRequiresManualReview());

        verify(notificationActivity, times(1))
                .sendCustomerNotification(eq("CUST-001"), eq("APP-001"), eq(KycStatus.APPROVED), anyString());
    }

    @Test
    @DisplayName("Should auto-reject applicant when identity document is expired or forged")
    void testDocumentFailureRejection() {
        KycWorkflowRequest request = createStandardRequest("APP-002", "CUST-002");

        when(documentActivity.verifyDocument(any()))
                .thenReturn(new DocumentVerificationResult(false, "Document has expired", 10.0, "E12345678"));

        WorkflowOptions options = WorkflowOptions.newBuilder()
                .setTaskQueue(TASK_QUEUE)
                .setWorkflowId("wf-test-doc-fail")
                .build();

        KycOnboardingWorkflow workflow = testEnv.getWorkflowClient().newWorkflowStub(KycOnboardingWorkflow.class, options);
        KycWorkflowResult result = workflow.executeKyc(request);

        assertNotNull(result);
        assertEquals(KycStatus.REJECTED, result.getFinalStatus());
        assertEquals("Document has expired", result.getRejectionReason());

        verify(biometricActivity, never()).verifyBiometrics(any());
        verify(sanctionsActivity, never()).screenSanctions(any());
    }

    @Test
    @DisplayName("Should auto-reject applicant when on sanctions/AML watchlist")
    void testSanctionsRejection() {
        KycWorkflowRequest request = createStandardRequest("APP-003", "CUST-003");

        when(documentActivity.verifyDocument(any()))
                .thenReturn(new DocumentVerificationResult(true, null, 98.0, "E12345678"));
        when(biometricActivity.verifyBiometrics(any()))
                .thenReturn(new BiometricVerificationResult(true, 90.0, true, "Face match high"));
        when(sanctionsActivity.screenSanctions(any()))
                .thenReturn(new SanctionsScreeningResult(false, false, true, "OFAC Blocked Persons"));

        WorkflowOptions options = WorkflowOptions.newBuilder()
                .setTaskQueue(TASK_QUEUE)
                .setWorkflowId("wf-test-sanctions-fail")
                .build();

        KycOnboardingWorkflow workflow = testEnv.getWorkflowClient().newWorkflowStub(KycOnboardingWorkflow.class, options);
        KycWorkflowResult result = workflow.executeKyc(request);

        assertNotNull(result);
        assertEquals(KycStatus.REJECTED, result.getFinalStatus());
        assertTrue(result.getRejectionReason().contains("Sanctions / AML watchlist match"));
    }

    @Test
    @DisplayName("Should pause for manual review when biometric is borderline and approve when signal arrives")
    void testManualReviewApprovalSignal() throws Exception {
        KycWorkflowRequest request = createStandardRequest("APP-004", "CUST-004");

        when(documentActivity.verifyDocument(any()))
                .thenReturn(new DocumentVerificationResult(true, null, 95.0, "E12345678"));
        // Borderline score: 72% (falls in 60-80% manual review bracket)
        when(biometricActivity.verifyBiometrics(any()))
                .thenReturn(new BiometricVerificationResult(true, 72.0, true, "Borderline match"));
        when(sanctionsActivity.screenSanctions(any()))
                .thenReturn(new SanctionsScreeningResult(true, false, false, "None"));
        when(riskActivity.assessRisk(any(), any(), any()))
                .thenReturn(new RiskAssessmentResult(45, true, false, "MEDIUM", "Borderline score"));

        WorkflowOptions options = WorkflowOptions.newBuilder()
                .setTaskQueue(TASK_QUEUE)
                .setWorkflowId("wf-test-manual-review")
                .build();

        KycOnboardingWorkflow workflow = testEnv.getWorkflowClient().newWorkflowStub(KycOnboardingWorkflow.class, options);

        // Start workflow asynchronously using Temporal client
        WorkflowClient.start(workflow::executeKyc, request);

        // Wait until workflow enters manual review state
        for (int i = 0; i < 30; i++) {
            if ("REQUIRES_MANUAL_REVIEW".equals(workflow.getWorkflowStatus())) {
                break;
            }
            Thread.sleep(100);
        }
        assertEquals("REQUIRES_MANUAL_REVIEW", workflow.getWorkflowStatus());

        // Send manual approval signal
        workflow.submitManualReview(new ManualReviewDecision("OFFICER-007", true, "Secondary identity proof verified"));

        // Wait for workflow completion
        KycWorkflowResult result = io.temporal.client.WorkflowStub.fromTyped(workflow).getResult(KycWorkflowResult.class);
        assertNotNull(result);
        assertEquals(KycStatus.APPROVED, result.getFinalStatus());
        assertEquals(WalletTier.TIER_1_BASIC, result.getWalletTier());
        assertTrue(result.isRequiresManualReview());

        verify(persistenceActivity, times(1))
                .recordManualReview(eq("APP-004"), eq("OFFICER-007"), eq("APPROVE"), anyString());
    }

    @Test
    @DisplayName("Should pause for manual review and reject when compliance officer denies")
    void testManualReviewRejectionSignal() throws Exception {
        KycWorkflowRequest request = createStandardRequest("APP-005", "CUST-005");

        when(documentActivity.verifyDocument(any()))
                .thenReturn(new DocumentVerificationResult(true, null, 95.0, "E12345678"));
        when(biometricActivity.verifyBiometrics(any()))
                .thenReturn(new BiometricVerificationResult(true, 70.0, true, "Borderline match"));
        when(sanctionsActivity.screenSanctions(any()))
                .thenReturn(new SanctionsScreeningResult(true, false, false, "None"));
        when(riskActivity.assessRisk(any(), any(), any()))
                .thenReturn(new RiskAssessmentResult(50, true, false, "MEDIUM", "Borderline"));

        WorkflowOptions options = WorkflowOptions.newBuilder()
                .setTaskQueue(TASK_QUEUE)
                .setWorkflowId("wf-test-manual-reject")
                .build();

        KycOnboardingWorkflow workflow = testEnv.getWorkflowClient().newWorkflowStub(KycOnboardingWorkflow.class, options);

        WorkflowClient.start(workflow::executeKyc, request);

        for (int i = 0; i < 30; i++) {
            if ("REQUIRES_MANUAL_REVIEW".equals(workflow.getWorkflowStatus())) {
                break;
            }
            Thread.sleep(100);
        }
        assertEquals("REQUIRES_MANUAL_REVIEW", workflow.getWorkflowStatus());

        workflow.submitManualReview(new ManualReviewDecision("OFFICER-007", false, "Photo does not match applicant"));

        KycWorkflowResult result = io.temporal.client.WorkflowStub.fromTyped(workflow).getResult(KycWorkflowResult.class);
        assertNotNull(result);
        assertEquals(KycStatus.REJECTED, result.getFinalStatus());
        assertTrue(result.getRejectionReason().contains("Photo does not match applicant"));
    }
}
