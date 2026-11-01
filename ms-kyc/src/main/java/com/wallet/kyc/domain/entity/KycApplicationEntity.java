package com.wallet.kyc.domain.entity;

import com.wallet.kyc.domain.DocumentType;
import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.domain.WalletTier;
import jakarta.persistence.*;
import java.time.LocalDate;
import java.time.LocalDateTime;

@Entity
@Table(name = "kyc_applications", indexes = {
    @Index(name = "idx_kyc_customer_id", columnList = "customer_id"),
    @Index(name = "idx_kyc_workflow_id", columnList = "workflow_id"),
    @Index(name = "idx_kyc_status", columnList = "status")
})
public class KycApplicationEntity {

    @Id
    @Column(name = "id", length = 64, nullable = false)
    private String id;

    @Column(name = "customer_id", length = 64, nullable = false)
    private String customerId;

    @Column(name = "workflow_id", length = 128)
    private String workflowId;

    @Column(name = "full_name", length = 128, nullable = false)
    private String fullName;

    @Column(name = "date_of_birth", nullable = false)
    private LocalDate dateOfBirth;

    @Column(name = "nationality", length = 8, nullable = false)
    private String nationality;

    @Column(name = "phone_number", length = 32, nullable = false)
    private String phoneNumber;

    @Column(name = "email", length = 128, nullable = false)
    private String email;

    @Enumerated(EnumType.STRING)
    @Column(name = "document_type", length = 32, nullable = false)
    private DocumentType documentType;

    @Column(name = "document_number", length = 64, nullable = false)
    private String documentNumber;

    @Column(name = "front_document_url", length = 512)
    private String frontDocumentUrl;

    @Column(name = "back_document_url", length = 512)
    private String backDocumentUrl;

    @Column(name = "selfie_url", length = 512)
    private String selfieUrl;

    @Enumerated(EnumType.STRING)
    @Column(name = "status", length = 32, nullable = false)
    private KycStatus status;

    @Column(name = "risk_score")
    private Integer riskScore;

    @Enumerated(EnumType.STRING)
    @Column(name = "wallet_tier", length = 32)
    private WalletTier walletTier;

    @Column(name = "rejection_reason", length = 512)
    private String rejectionReason;

    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @Column(name = "updated_at", nullable = false)
    private LocalDateTime updatedAt;

    public KycApplicationEntity() {}

    @PrePersist
    public void prePersist() {
        LocalDateTime now = LocalDateTime.now();
        this.createdAt = now;
        this.updatedAt = now;
        if (this.status == null) {
            this.status = KycStatus.PENDING;
        }
        if (this.walletTier == null) {
            this.walletTier = WalletTier.NONE;
        }
    }

    @PreUpdate
    public void preUpdate() {
        this.updatedAt = LocalDateTime.now();
    }

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }

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

    public String getFrontDocumentUrl() { return frontDocumentUrl; }
    public void setFrontDocumentUrl(String frontDocumentUrl) { this.frontDocumentUrl = frontDocumentUrl; }

    public String getBackDocumentUrl() { return backDocumentUrl; }
    public void setBackDocumentUrl(String backDocumentUrl) { this.backDocumentUrl = backDocumentUrl; }

    public String getSelfieUrl() { return selfieUrl; }
    public void setSelfieUrl(String selfieUrl) { this.selfieUrl = selfieUrl; }

    public KycStatus getStatus() { return status; }
    public void setStatus(KycStatus status) { this.status = status; }

    public Integer getRiskScore() { return riskScore; }
    public void setRiskScore(Integer riskScore) { this.riskScore = riskScore; }

    public WalletTier getWalletTier() { return walletTier; }
    public void setWalletTier(WalletTier walletTier) { this.walletTier = walletTier; }

    public String getRejectionReason() { return rejectionReason; }
    public void setRejectionReason(String rejectionReason) { this.rejectionReason = rejectionReason; }

    public LocalDateTime getCreatedAt() { return createdAt; }
    public void setCreatedAt(LocalDateTime createdAt) { this.createdAt = createdAt; }

    public LocalDateTime getUpdatedAt() { return updatedAt; }
    public void setUpdatedAt(LocalDateTime updatedAt) { this.updatedAt = updatedAt; }
}
