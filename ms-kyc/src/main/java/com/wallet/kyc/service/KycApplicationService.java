package com.wallet.kyc.service;

import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.domain.WalletTier;
import com.wallet.kyc.domain.entity.KycApplicationEntity;
import com.wallet.kyc.domain.entity.KycManualReviewEntity;
import com.wallet.kyc.domain.entity.KycVerificationRecordEntity;
import com.wallet.kyc.dto.*;
import com.wallet.kyc.repository.KycApplicationRepository;
import com.wallet.kyc.repository.KycManualReviewRepository;
import com.wallet.kyc.repository.KycVerificationRecordRepository;
import com.wallet.kyc.temporal.model.KycWorkflowProgress;
import com.wallet.kyc.temporal.model.KycWorkflowRequest;
import com.wallet.kyc.temporal.model.ManualReviewDecision;
import com.wallet.kyc.temporal.workflow.KycOnboardingWorkflow;
import io.temporal.api.enums.v1.WorkflowIdReusePolicy;
import io.temporal.client.WorkflowClient;
import io.temporal.client.WorkflowOptions;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.slf4j.MDC;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
public class KycApplicationService {

    private static final Logger log = LoggerFactory.getLogger(KycApplicationService.class);

    private final KycApplicationRepository applicationRepository;
    private final KycVerificationRecordRepository recordRepository;
    private final KycManualReviewRepository manualReviewRepository;
    private final WorkflowClient workflowClient;

    @Value("${temporal.task-queue:KYC_ONBOARDING_TASK_QUEUE}")
    private String taskQueue;

    public KycApplicationService(KycApplicationRepository applicationRepository,
                                 KycVerificationRecordRepository recordRepository,
                                 KycManualReviewRepository manualReviewRepository,
                                 WorkflowClient workflowClient) {
        this.applicationRepository = applicationRepository;
        this.recordRepository = recordRepository;
        this.manualReviewRepository = manualReviewRepository;
        this.workflowClient = workflowClient;
    }

    @Transactional
    public KycApplicationResponseDto initiateOnboarding(KycOnboardingRequestDto requestDto) {
        if (requestDto.getCustomerId() != null) {
            MDC.put("customer_id", requestDto.getCustomerId());
        }
        String datePrefix = LocalDate.now().format(DateTimeFormatter.ofPattern("yyyyMMdd"));
        String randomSuffix = UUID.randomUUID().toString().substring(0, 8).toUpperCase();
        String applicationId = "KYC-" + datePrefix + "-" + randomSuffix;
        String workflowId = "kyc-wf-" + requestDto.getCustomerId() + "-" + randomSuffix;

        log.info("Initiating onboarding for customerId={}, applicationId={}, workflowId={}",
                requestDto.getCustomerId(), applicationId, workflowId);

        KycApplicationEntity entity = new KycApplicationEntity();
        entity.setId(applicationId);
        entity.setCustomerId(requestDto.getCustomerId());
        entity.setWorkflowId(workflowId);
        entity.setFullName(requestDto.getFullName());
        entity.setDateOfBirth(requestDto.getDateOfBirth());
        entity.setNationality(requestDto.getNationality());
        entity.setPhoneNumber(requestDto.getPhoneNumber());
        entity.setEmail(requestDto.getEmail());
        entity.setDocumentType(requestDto.getDocumentType());
        entity.setDocumentNumber(requestDto.getDocumentNumber());
        entity.setFrontDocumentUrl(requestDto.getFrontDocumentUrl());
        entity.setBackDocumentUrl(requestDto.getBackDocumentUrl());
        entity.setSelfieUrl(requestDto.getSelfieUrl());
        entity.setStatus(KycStatus.PENDING);
        entity.setWalletTier(WalletTier.NONE);

        KycApplicationEntity savedEntity = applicationRepository.save(entity);

        // Prepare Temporal Workflow Request
        KycWorkflowRequest workflowRequest = new KycWorkflowRequest(
                applicationId,
                requestDto.getCustomerId(),
                requestDto.getFullName(),
                requestDto.getDateOfBirth(),
                requestDto.getNationality(),
                requestDto.getPhoneNumber(),
                requestDto.getEmail(),
                requestDto.getDocumentType(),
                requestDto.getDocumentNumber(),
                requestDto.getFrontDocumentUrl(),
                requestDto.getBackDocumentUrl(),
                requestDto.getSelfieUrl()
        );

        WorkflowOptions options = WorkflowOptions.newBuilder()
                .setWorkflowId(workflowId)
                .setTaskQueue(taskQueue)
                .setWorkflowIdReusePolicy(WorkflowIdReusePolicy.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE)
                .build();

        KycOnboardingWorkflow workflow = workflowClient.newWorkflowStub(KycOnboardingWorkflow.class, options);
        WorkflowClient.start(workflow::executeKyc, workflowRequest);

        log.info("Started Temporal workflow {} for applicationId {}", workflowId, applicationId);

        return mapToResponseDto(savedEntity, List.of(), List.of());
    }

    public KycApplicationResponseDto getApplicationById(String applicationId) {
        KycApplicationEntity entity = applicationRepository.findById(applicationId)
                .orElseThrow(() -> new NoSuchElementException("KYC Application not found with id: " + applicationId));

        List<KycVerificationRecordEntity> records = recordRepository.findByApplicationIdOrderByCreatedAtAsc(applicationId);
        List<KycManualReviewEntity> reviews = manualReviewRepository.findByApplicationIdOrderByReviewedAtDesc(applicationId);

        return mapToResponseDto(entity, records, reviews);
    }

    public KycApplicationResponseDto getLatestByCustomerId(String customerId) {
        KycApplicationEntity entity = applicationRepository.findFirstByCustomerIdOrderByCreatedAtDesc(customerId)
                .orElseThrow(() -> new NoSuchElementException("No KYC Application found for customer: " + customerId));

        List<KycVerificationRecordEntity> records = recordRepository.findByApplicationIdOrderByCreatedAtAsc(entity.getId());
        List<KycManualReviewEntity> reviews = manualReviewRepository.findByApplicationIdOrderByReviewedAtDesc(entity.getId());

        return mapToResponseDto(entity, records, reviews);
    }

    public List<KycApplicationResponseDto> listApplications(KycStatus status) {
        List<KycApplicationEntity> list = (status == null)
                ? applicationRepository.findAll()
                : applicationRepository.findByStatusOrderByCreatedAtDesc(status);

        return list.stream()
                .map(entity -> mapToResponseDto(entity, List.of(), List.of()))
                .collect(Collectors.toList());
    }

    public WorkflowProgressResponseDto getWorkflowProgress(String applicationId) {
        KycApplicationEntity entity = applicationRepository.findById(applicationId)
                .orElseThrow(() -> new NoSuchElementException("KYC Application not found with id: " + applicationId));

        try {
            KycOnboardingWorkflow workflow = workflowClient.newWorkflowStub(
                    KycOnboardingWorkflow.class, entity.getWorkflowId());
            KycWorkflowProgress progress = workflow.getProgress();
            return new WorkflowProgressResponseDto(
                    applicationId,
                    entity.getWorkflowId(),
                    progress.getCurrentStage(),
                    progress.getPercentComplete(),
                    progress.getStatusMessage()
            );
        } catch (Exception e) {
            log.warn("Could not query Temporal workflow directly (falling back to database state): {}", e.getMessage());
            int percent = (entity.getStatus() == KycStatus.APPROVED || entity.getStatus() == KycStatus.REJECTED) ? 100 : 50;
            return new WorkflowProgressResponseDto(
                    applicationId,
                    entity.getWorkflowId(),
                    entity.getStatus().name(),
                    percent,
                    "Application status: " + entity.getStatus()
            );
        }
    }

    public void submitManualReview(String applicationId, ManualReviewRequestDto reviewDto) {
        KycApplicationEntity entity = applicationRepository.findById(applicationId)
                .orElseThrow(() -> new NoSuchElementException("KYC Application not found with id: " + applicationId));

        if (entity.getStatus() != KycStatus.REQUIRES_MANUAL_REVIEW) {
            throw new IllegalStateException("Application is not in REQUIRES_MANUAL_REVIEW status. Current status: " + entity.getStatus());
        }

        boolean approved = "APPROVE".equalsIgnoreCase(reviewDto.getDecision());
        ManualReviewDecision decision = new ManualReviewDecision(
                reviewDto.getReviewerId(),
                approved,
                reviewDto.getNotes()
        );

        log.info("Sending manual review signal to workflow {} for applicationId {}", entity.getWorkflowId(), applicationId);
        KycOnboardingWorkflow workflow = workflowClient.newWorkflowStub(
                KycOnboardingWorkflow.class, entity.getWorkflowId());
        workflow.submitManualReview(decision);
    }

    private KycApplicationResponseDto mapToResponseDto(KycApplicationEntity entity,
                                                      List<KycVerificationRecordEntity> records,
                                                      List<KycManualReviewEntity> reviews) {
        KycApplicationResponseDto dto = new KycApplicationResponseDto();
        dto.setApplicationId(entity.getId());
        dto.setCustomerId(entity.getCustomerId());
        dto.setWorkflowId(entity.getWorkflowId());
        dto.setFullName(entity.getFullName());
        dto.setDateOfBirth(entity.getDateOfBirth());
        dto.setNationality(entity.getNationality());
        dto.setPhoneNumber(entity.getPhoneNumber());
        dto.setEmail(entity.getEmail());
        dto.setDocumentType(entity.getDocumentType());
        dto.setDocumentNumber(entity.getDocumentNumber());
        dto.setStatus(entity.getStatus());
        dto.setWalletTier(entity.getWalletTier());
        dto.setRiskScore(entity.getRiskScore());
        dto.setRejectionReason(entity.getRejectionReason());
        dto.setCreatedAt(entity.getCreatedAt());
        dto.setUpdatedAt(entity.getUpdatedAt());

        dto.setVerificationRecords(records.stream()
                .map(r -> new KycApplicationResponseDto.VerificationRecordDto(
                        r.getId(), r.getStepName(), r.isPassed(), r.getScore(), r.getDetailsJson(), r.getCreatedAt()))
                .collect(Collectors.toList()));

        dto.setManualReviews(reviews.stream()
                .map(m -> new KycApplicationResponseDto.ManualReviewDto(
                        m.getId(), m.getReviewerId(), m.getDecision(), m.getNotes(), m.getReviewedAt()))
                .collect(Collectors.toList()));

        return dto;
    }
}
