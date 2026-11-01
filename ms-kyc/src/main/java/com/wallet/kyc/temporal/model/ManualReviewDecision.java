package com.wallet.kyc.temporal.model;

import java.io.Serializable;
import java.time.LocalDateTime;

public class ManualReviewDecision implements Serializable {
    private String reviewerId;
    private boolean approved;
    private String notes;
    private LocalDateTime decisionTime;

    public ManualReviewDecision() {}

    public ManualReviewDecision(String reviewerId, boolean approved, String notes) {
        this.reviewerId = reviewerId;
        this.approved = approved;
        this.notes = notes;
        this.decisionTime = LocalDateTime.now();
    }

    public String getReviewerId() { return reviewerId; }
    public void setReviewerId(String reviewerId) { this.reviewerId = reviewerId; }

    public boolean isApproved() { return approved; }
    public void setApproved(boolean approved) { this.approved = approved; }

    public String getNotes() { return notes; }
    public void setNotes(String notes) { this.notes = notes; }

    public LocalDateTime getDecisionTime() { return decisionTime; }
    public void setDecisionTime(LocalDateTime decisionTime) { this.decisionTime = decisionTime; }
}
