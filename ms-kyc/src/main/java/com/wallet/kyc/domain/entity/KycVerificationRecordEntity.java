package com.wallet.kyc.domain.entity;

import jakarta.persistence.*;
import java.time.LocalDateTime;

@Entity
@Table(name = "kyc_verification_records", indexes = {
    @Index(name = "idx_kvr_app_id", columnList = "application_id")
})
public class KycVerificationRecordEntity {

    @Id
    @Column(name = "id", length = 64, nullable = false)
    private String id;

    @Column(name = "application_id", length = 64, nullable = false)
    private String applicationId;

    @Column(name = "step_name", length = 64, nullable = false)
    private String stepName;

    @Column(name = "passed", nullable = false)
    private boolean passed;

    @Column(name = "score")
    private Double score;

    @Column(name = "details_json", length = 2048)
    private String detailsJson;

    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    public KycVerificationRecordEntity() {}

    public KycVerificationRecordEntity(String id, String applicationId, String stepName, boolean passed, Double score, String detailsJson) {
        this.id = id;
        this.applicationId = applicationId;
        this.stepName = stepName;
        this.passed = passed;
        this.score = score;
        this.detailsJson = detailsJson;
        this.createdAt = LocalDateTime.now();
    }

    @PrePersist
    public void prePersist() {
        if (this.createdAt == null) {
            this.createdAt = LocalDateTime.now();
        }
    }

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }

    public String getApplicationId() { return applicationId; }
    public void setApplicationId(String applicationId) { this.applicationId = applicationId; }

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
