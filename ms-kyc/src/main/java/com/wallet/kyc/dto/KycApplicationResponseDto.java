package com.wallet.kyc.dto;

import com.wallet.kyc.domain.DocumentType;
import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.domain.WalletTier;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;

public class KycApplicationResponseDto {

    private String applicationId;
    private String customerId;
    private String workflowId;
    private String fullName;
    private LocalDate dateOfBirth;
    private String nationality;
    private String phoneNumber;
    private String email;
    private DocumentType documentType;
    private String documentNumber;
    private KycStatus status;
    private WalletTier walletTier;
    private Integer riskScore;
    private String rejectionReason;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
    private List<VerificationRecordDto> verificationRecords;
    private List<ManualReviewDto> manualReviews;

    public KycApplicationResponseDto() {}

    public static class VerificationRecordDto {
        private String id;
        private String stepName;
        private boolean passed;
        private Double score;
        private String detailsJson;
        private LocalDateTime createdAt;

        public VerificationRecordDto() {}

        public VerificationRecordDto(String id, String stepName, boolean passed, Double score, String detailsJson, LocalDateTime createdAt) {
            this.id = id;
            this.stepName = stepName;
            this.passed = passed;
            this.score = score;
            this.detailsJson = detailsJson;
            this.createdAt = createdAt;
        }

        public String getId() { return id; }
        public void setId(String id) { this.id = id; }
        public String getStepName() { return stepName; }
        public void setStepName(String stepName) { this.stepName = stepName; }
        public boolean isPassed() { return passed; }
        public void setPassed(boolean passed) { this.passed = passed; }
        public Double getScore() { return score; }
        public void setScore(Double score) { this.score = score; }
        public String getDetailsJson() { return detailsJson; }
        public void setDetailsJson(String detailsJson) { this.detailsJson = detailsJson; }
        public LocalDateTime getCreatedAt() { return createdAt; }
        public void setCreatedAt(LocalDateTime createdAt) { this.createdAt = createdAt; }
    }

    public static class ManualReviewDto {
        private String id;
        private String reviewerId;
        private String decision;
        private String notes;
        private LocalDateTime reviewedAt;

        public ManualReviewDto() {}

        public ManualReviewDto(String id, String reviewerId, String decision, String notes, LocalDateTime reviewedAt) {
            this.id = id;
            this.reviewerId = reviewerId;
            this.decision = decision;
            this.notes = notes;
            this.reviewedAt = reviewedAt;
        }

        public String getId() { return id; }
        public void setId(String id) { this.id = id; }
        public String getReviewerId() { return reviewerId; }
        public void setReviewerId(String reviewerId) { this.reviewerId = reviewerId; }
        public String getDecision() { return decision; }
        public void setDecision(String decision) { this.decision = decision; }
        public String getNotes() { return notes; }
        public void setNotes(String notes) { this.notes = notes; }
        public LocalDateTime getReviewedAt() { return reviewedAt; }
        public void setReviewedAt(LocalDateTime reviewedAt) { this.reviewedAt = reviewedAt; }
    }

    public String getApplicationId() { return applicationId; }
    public void setApplicationId(String applicationId) { this.applicationId = applicationId; }

    public String getCustomerId() { return customerId; }
    public void setCustomerId(String customerId) { this.customerId = customerId; }

    public String getWorkflowId() { return workflowId; }
    public void setWorkflowId(String workflowId) { this.workflowId = workflowId; }

    public String getFullName() { return fullName; }
    public void setFullName(String fullName) { this.fullName = fullName; }

    public LocalDate getDateOfBirth() { return dateOfBirth; }
    public void setDateOfBirth(LocalDate dateOfBirth) { this.dateOfBirth = dateOfBirth; }

    public String getNationality() { return nationality; }
    public void setNationality(String nationality) { this.nationality = nationality; }

    public String getPhoneNumber() { return phoneNumber; }
    public void setPhoneNumber(String phoneNumber) { this.phoneNumber = phoneNumber; }

    public String getEmail() { return email; }
    public void setEmail(String email) { this.email = email; }

    public DocumentType getDocumentType() { return documentType; }
    public void setDocumentType(DocumentType documentType) { this.documentType = documentType; }

    public String getDocumentNumber() { return documentNumber; }
    public void setDocumentNumber(String documentNumber) { this.documentNumber = documentNumber; }

    public KycStatus getStatus() { return status; }
    public void setStatus(KycStatus status) { this.status = status; }

    public WalletTier getWalletTier() { return walletTier; }
    public void setWalletTier(WalletTier walletTier) { this.walletTier = walletTier; }

    public Integer getRiskScore() { return riskScore; }
    public void setRiskScore(Integer riskScore) { this.riskScore = riskScore; }

    public String getRejectionReason() { return rejectionReason; }
    public void setRejectionReason(String rejectionReason) { this.rejectionReason = rejectionReason; }

    public LocalDateTime getCreatedAt() { return createdAt; }
    public void setCreatedAt(LocalDateTime createdAt) { this.createdAt = createdAt; }

    public LocalDateTime getUpdatedAt() { return updatedAt; }
    public void setUpdatedAt(LocalDateTime updatedAt) { this.updatedAt = updatedAt; }

    public List<VerificationRecordDto> getVerificationRecords() { return verificationRecords; }
    public void setVerificationRecords(List<VerificationRecordDto> verificationRecords) { this.verificationRecords = verificationRecords; }

    public List<ManualReviewDto> getManualReviews() { return manualReviews; }
    public void setManualReviews(List<ManualReviewDto> manualReviews) { this.manualReviews = manualReviews; }
}
