package com.wallet.kyc.domain.entity;

import jakarta.persistence.*;
import java.time.LocalDateTime;

@Entity
@Table(name = "kyc_manual_reviews", indexes = {
    @Index(name = "idx_kmr_app_id", columnList = "application_id")
})
public class KycManualReviewEntity {

    @Id
    @Column(name = "id", length = 64, nullable = false)
    private String id;

    @Column(name = "application_id", length = 64, nullable = false)
    private String applicationId;

    @Column(name = "reviewer_id", length = 64, nullable = false)
    private String reviewerId;

    @Column(name = "decision", length = 32, nullable = false)
    private String decision;

    @Column(name = "notes", length = 1024)
    private String notes;

    @Column(name = "reviewed_at", nullable = false)
    private LocalDateTime reviewedAt;

    public KycManualReviewEntity() {}

    public KycManualReviewEntity(String id, String applicationId, String reviewerId, String decision, String notes) {
        this.id = id;
        this.applicationId = applicationId;
        this.reviewerId = reviewerId;
        this.decision = decision;
        this.notes = notes;
        this.reviewedAt = LocalDateTime.now();
    }

    @PrePersist
    public void prePersist() {
        if (this.reviewedAt == null) {
            this.reviewedAt = LocalDateTime.now();
        }
    }

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }

    public String getApplicationId() { return applicationId; }
    public void setApplicationId(String applicationId) { this.applicationId = applicationId; }

    public String getReviewerId() { return reviewerId; }
    public void setReviewerId(String reviewerId) { this.reviewerId = reviewerId; }

    public String getDecision() { return decision; }
    public void setDecision(String decision) { this.decision = decision; }

    public String getNotes() { return notes; }
    public void setNotes(String notes) { this.notes = notes; }

    public LocalDateTime getReviewedAt() { return reviewedAt; }
    public void setReviewedAt(LocalDateTime reviewedAt) { this.reviewedAt = reviewedAt; }
}
