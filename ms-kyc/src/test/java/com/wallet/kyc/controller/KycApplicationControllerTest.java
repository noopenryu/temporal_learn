package com.wallet.kyc.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.wallet.kyc.domain.DocumentType;
import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.domain.WalletTier;
import com.wallet.kyc.dto.KycApplicationResponseDto;
import com.wallet.kyc.dto.KycOnboardingRequestDto;
import com.wallet.kyc.dto.ManualReviewRequestDto;
import com.wallet.kyc.dto.WorkflowProgressResponseDto;
import com.wallet.kyc.service.KycApplicationService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;

import java.time.LocalDate;
import java.util.NoSuchElementException;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@ExtendWith(MockitoExtension.class)
class KycApplicationControllerTest {

    private MockMvc mockMvc;
    private ObjectMapper objectMapper;

    @Mock
    private KycApplicationService applicationService;

    @BeforeEach
    void setUp() {
        objectMapper = new ObjectMapper();
        objectMapper.registerModule(new JavaTimeModule());

        KycApplicationController controller = new KycApplicationController(applicationService);
        mockMvc = MockMvcBuilders.standaloneSetup(controller)
                .setControllerAdvice(new GlobalExceptionHandler())
                .build();
    }

    @Test
    @DisplayName("POST /api/v1/kyc/onboard should return 202 ACCEPTED when payload is valid")
    void testInitiateOnboardingSuccess() throws Exception {
        KycOnboardingRequestDto request = new KycOnboardingRequestDto();
        request.setCustomerId("CUST-1001");
        request.setFullName("John Doe");
        request.setDateOfBirth(LocalDate.of(1990, 1, 1));
        request.setNationality("SGP");
        request.setPhoneNumber("+6591234567");
        request.setEmail("john@example.com");
        request.setDocumentType(DocumentType.PASSPORT);
        request.setDocumentNumber("E12345678");
        request.setFrontDocumentUrl("https://cdn.example.com/front.jpg");
        request.setSelfieUrl("https://cdn.example.com/selfie.jpg");

        KycApplicationResponseDto responseDto = new KycApplicationResponseDto();
        responseDto.setApplicationId("KYC-20260927-ABC12345");
        responseDto.setCustomerId("CUST-1001");
        responseDto.setStatus(KycStatus.PENDING);
        responseDto.setWalletTier(WalletTier.NONE);

        when(applicationService.initiateOnboarding(any())).thenReturn(responseDto);

        mockMvc.perform(post("/api/v1/kyc/onboard")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isAccepted())
                .andExpect(jsonPath("$.applicationId").value("KYC-20260927-ABC12345"))
                .andExpect(jsonPath("$.customerId").value("CUST-1001"))
                .andExpect(jsonPath("$.status").value("PENDING"));

        verify(applicationService, times(1)).initiateOnboarding(any());
    }

    @Test
    @DisplayName("POST /api/v1/kyc/onboard should return 400 BAD REQUEST when required fields are missing")
    void testInitiateOnboardingValidationFailure() throws Exception {
        KycOnboardingRequestDto request = new KycOnboardingRequestDto();
        request.setEmail("invalid-email-format");

        mockMvc.perform(post("/api/v1/kyc/onboard")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.error").value("Validation Failed"))
                .andExpect(jsonPath("$.details.customerId").exists());

        verify(applicationService, never()).initiateOnboarding(any());
    }

    @Test
    @DisplayName("GET /api/v1/kyc/applications/{id} should return 200 OK")
    void testGetApplicationSuccess() throws Exception {
        String appId = "KYC-20260927-TEST001";
        KycApplicationResponseDto responseDto = new KycApplicationResponseDto();
        responseDto.setApplicationId(appId);
        responseDto.setStatus(KycStatus.APPROVED);
        responseDto.setWalletTier(WalletTier.TIER_2_PREMIUM);

        when(applicationService.getApplicationById(appId)).thenReturn(responseDto);

        mockMvc.perform(get("/api/v1/kyc/applications/" + appId))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.applicationId").value(appId))
                .andExpect(jsonPath("$.status").value("APPROVED"))
                .andExpect(jsonPath("$.walletTier").value("TIER_2_PREMIUM"));
    }

    @Test
    @DisplayName("GET /api/v1/kyc/applications/{id} should return 404 NOT FOUND when application does not exist")
    void testGetApplicationNotFound() throws Exception {
        String appId = "KYC-NONEXISTENT";
        when(applicationService.getApplicationById(appId))
                .thenThrow(new NoSuchElementException("KYC Application not found with id: " + appId));

        mockMvc.perform(get("/api/v1/kyc/applications/" + appId))
                .andExpect(status().isNotFound())
                .andExpect(jsonPath("$.error").value("Not Found"));
    }

    @Test
    @DisplayName("POST /api/v1/kyc/applications/{id}/manual-review should return 200 OK")
    void testSubmitManualReviewSuccess() throws Exception {
        String appId = "KYC-20260927-REV001";
        ManualReviewRequestDto request = new ManualReviewRequestDto("OFFICER-007", "APPROVE", "Identity documents verified");

        doNothing().when(applicationService).submitManualReview(eq(appId), any());

        mockMvc.perform(post("/api/v1/kyc/applications/" + appId + "/manual-review")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.decision").value("APPROVE"))
                .andExpect(jsonPath("$.message").value("Manual review signal submitted successfully"));

        verify(applicationService, times(1)).submitManualReview(eq(appId), any());
    }

    @Test
    @DisplayName("GET /api/v1/kyc/applications/{id}/progress should return 200 OK with workflow progress")
    void testGetProgressSuccess() throws Exception {
        String appId = "KYC-20260927-PROG001";
        WorkflowProgressResponseDto progress = new WorkflowProgressResponseDto(
                appId, "kyc-wf-001", "BIOMETRIC_MATCH", 35, "Analyzing face match"
        );

        when(applicationService.getWorkflowProgress(appId)).thenReturn(progress);

        mockMvc.perform(get("/api/v1/kyc/applications/" + appId + "/progress"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.applicationId").value(appId))
                .andExpect(jsonPath("$.currentStage").value("BIOMETRIC_MATCH"))
                .andExpect(jsonPath("$.percentComplete").value(35));
    }
}
