package com.wallet.kyc.temporal.activity;

import com.wallet.kyc.domain.DocumentType;
import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.repository.KycApplicationRepository;
import com.wallet.kyc.temporal.activity.impl.BiometricVerificationActivityImpl;
import com.wallet.kyc.temporal.activity.impl.DocumentVerificationActivityImpl;
import com.wallet.kyc.temporal.activity.impl.RiskAssessmentActivityImpl;
import com.wallet.kyc.temporal.activity.impl.SanctionsScreeningActivityImpl;
import com.wallet.kyc.temporal.model.*;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import java.time.LocalDate;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.when;

class ActivityUnitTests {

    private KycWorkflowRequest createRequest(String docNumber, String name, String nationality, LocalDate dob, String selfieUrl, String frontDocUrl) {
        return new KycWorkflowRequest(
                "APP-100",
                "CUST-100",
                name,
                dob,
                nationality,
                "+6591234567",
                "test@example.com",
                DocumentType.PASSPORT,
                docNumber,
                frontDocUrl,
                null,
                selfieUrl
        );
    }

    @Test
    @DisplayName("DocumentVerificationActivity: should validate document number format and forgery")
    void testDocumentVerification() {
        DocumentVerificationActivityImpl activity = new DocumentVerificationActivityImpl();

        // Valid
        KycWorkflowRequest validReq = createRequest("P12345678", "Bob", "SGP", LocalDate.of(1990, 1, 1), "http://selfie", "http://doc");
        DocumentVerificationResult r1 = activity.verifyDocument(validReq);
        assertTrue(r1.isValid());
        assertEquals("P12345678", r1.getExtractedDocumentNumber());

        // Forged flag
        KycWorkflowRequest forgedReq = createRequest("FORGED999", "Bob", "SGP", LocalDate.of(1990, 1, 1), "http://selfie", "http://doc");
        DocumentVerificationResult r2 = activity.verifyDocument(forgedReq);
        assertFalse(r2.isValid());
        assertTrue(r2.getFailureReason().contains("forgery"));

        // Missing front image
        KycWorkflowRequest missingImg = createRequest("P12345678", "Bob", "SGP", LocalDate.of(1990, 1, 1), "http://selfie", null);
        DocumentVerificationResult r3 = activity.verifyDocument(missingImg);
        assertFalse(r3.isValid());
    }

    @Test
    @DisplayName("BiometricVerificationActivity: should return match scores and detect mismatch / borderline")
    void testBiometricVerification() {
        BiometricVerificationActivityImpl activity = new BiometricVerificationActivityImpl();

        // Normal match
        KycWorkflowRequest normal = createRequest("P123", "Bob", "SGP", LocalDate.of(1990, 1, 1), "https://cdn/alice.jpg", "https://cdn/doc.jpg");
        BiometricVerificationResult r1 = activity.verifyBiometrics(normal);
        assertTrue(r1.isMatched());
        assertTrue(r1.getFaceMatchScore() > 80.0);

        // Borderline flag
        KycWorkflowRequest borderline = createRequest("P123", "Bob", "SGP", LocalDate.of(1990, 1, 1), "https://cdn/alice-borderline.jpg", "https://cdn/doc.jpg");
        BiometricVerificationResult r2 = activity.verifyBiometrics(borderline);
        assertTrue(r2.isMatched());
        assertEquals(72.5, r2.getFaceMatchScore());

        // Mismatch flag
        KycWorkflowRequest mismatch = createRequest("P123", "Bob", "SGP", LocalDate.of(1990, 1, 1), "https://cdn/alice-mismatch.jpg", "https://cdn/doc.jpg");
        BiometricVerificationResult r3 = activity.verifyBiometrics(mismatch);
        assertFalse(r3.isMatched());
        assertTrue(r3.getFaceMatchScore() < 60.0);
    }

    @Test
    @DisplayName("SanctionsScreeningActivity: should detect sanctions, PEP, and FATF blacklists")
    void testSanctionsScreening() {
        SanctionsScreeningActivityImpl activity = new SanctionsScreeningActivityImpl();

        // Clean
        KycWorkflowRequest clean = createRequest("P123", "Charlie Green", "SGP", LocalDate.of(1990, 1, 1), "http://selfie", "http://doc");
        SanctionsScreeningResult r1 = activity.screenSanctions(clean);
        assertTrue(r1.isClear());

        // Sanctioned
        KycWorkflowRequest sanctioned = createRequest("P123", "Mr Sanction Target", "SGP", LocalDate.of(1990, 1, 1), "http://selfie", "http://doc");
        SanctionsScreeningResult r2 = activity.screenSanctions(sanctioned);
        assertFalse(r2.isClear());
        assertTrue(r2.isSanctionsMatch());

        // High risk country
        KycWorkflowRequest highRisk = createRequest("P123", "Charlie Green", "NK", LocalDate.of(1990, 1, 1), "http://selfie", "http://doc");
        SanctionsScreeningResult r3 = activity.screenSanctions(highRisk);
        assertFalse(r3.isClear());
    }

    @Test
    @DisplayName("RiskAssessmentActivity: should enforce age limit and compute risk score")
    void testRiskAssessment() {
        KycApplicationRepository repo = Mockito.mock(KycApplicationRepository.class);
        when(repo.existsByDocumentTypeAndDocumentNumberAndStatus(any(), any(), eq(KycStatus.APPROVED)))
                .thenReturn(false);

        RiskAssessmentActivityImpl activity = new RiskAssessmentActivityImpl(repo);

        // Adult applicant (30 years old)
        KycWorkflowRequest adult = createRequest("P123", "David", "SGP", LocalDate.now().minusYears(30), "http://selfie", "http://doc");
        DocumentVerificationResult doc = new DocumentVerificationResult(true, null, 99.0, "P123");
        BiometricVerificationResult bio = new BiometricVerificationResult(true, 95.0, true, "OK");

        RiskAssessmentResult r1 = activity.assessRisk(adult, doc, bio);
        assertTrue(r1.isAgeEligible());
        assertFalse(r1.isDuplicateDetected());
        assertEquals("LOW", r1.getRiskLevel());

        // Underage applicant (16 years old)
        KycWorkflowRequest underage = createRequest("P123", "David Jr", "SGP", LocalDate.now().minusYears(16), "http://selfie", "http://doc");
        RiskAssessmentResult r2 = activity.assessRisk(underage, doc, bio);
        assertFalse(r2.isAgeEligible());
    }
}
