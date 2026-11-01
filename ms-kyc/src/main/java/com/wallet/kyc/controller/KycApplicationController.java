package com.wallet.kyc.controller;

import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.dto.KycApplicationResponseDto;
import com.wallet.kyc.dto.KycOnboardingRequestDto;
import com.wallet.kyc.dto.ManualReviewRequestDto;
import com.wallet.kyc.dto.WorkflowProgressResponseDto;
import com.wallet.kyc.service.KycApplicationService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@CrossOrigin(origins = "*")
@RequestMapping("/api/v1/kyc")
public class KycApplicationController {

    private final KycApplicationService applicationService;

    public KycApplicationController(KycApplicationService applicationService) {
        this.applicationService = applicationService;
    }

    @PostMapping("/onboard")
    public ResponseEntity<KycApplicationResponseDto> initiateOnboarding(
            @Valid @RequestBody KycOnboardingRequestDto requestDto) {
        KycApplicationResponseDto response = applicationService.initiateOnboarding(requestDto);
        return ResponseEntity.status(HttpStatus.ACCEPTED).body(response);
    }

    @GetMapping("/applications/{applicationId}")
    public ResponseEntity<KycApplicationResponseDto> getApplication(
            @PathVariable String applicationId) {
        KycApplicationResponseDto response = applicationService.getApplicationById(applicationId);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/applications/{applicationId}/progress")
    public ResponseEntity<WorkflowProgressResponseDto> getWorkflowProgress(
            @PathVariable String applicationId) {
        WorkflowProgressResponseDto progress = applicationService.getWorkflowProgress(applicationId);
        return ResponseEntity.ok(progress);
    }

    @PostMapping("/applications/{applicationId}/manual-review")
    public ResponseEntity<Map<String, String>> submitManualReview(
            @PathVariable String applicationId,
            @Valid @RequestBody ManualReviewRequestDto reviewDto) {
        applicationService.submitManualReview(applicationId, reviewDto);
        return ResponseEntity.ok(Map.of(
                "applicationId", applicationId,
                "message", "Manual review signal submitted successfully",
                "decision", reviewDto.getDecision()
        ));
    }

    @GetMapping("/customers/{customerId}")
    public ResponseEntity<KycApplicationResponseDto> getCustomerKyc(
            @PathVariable String customerId) {
        KycApplicationResponseDto response = applicationService.getLatestByCustomerId(customerId);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/applications")
    public ResponseEntity<List<KycApplicationResponseDto>> listApplications(
            @RequestParam(required = false) KycStatus status) {
        List<KycApplicationResponseDto> list = applicationService.listApplications(status);
        return ResponseEntity.ok(list);
    }
}
