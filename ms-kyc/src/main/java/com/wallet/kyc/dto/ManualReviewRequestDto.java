package com.wallet.kyc.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;

public class ManualReviewRequestDto {

    @NotBlank(message = "reviewerId is required")
    private String reviewerId;

    @NotBlank(message = "decision is required")
    @Pattern(regexp = "APPROVE|REJECT", message = "decision must be either APPROVE or REJECT")
    private String decision;

    @NotBlank(message = "notes is required")
    private String notes;

    public ManualReviewRequestDto() {}

    public ManualReviewRequestDto(String reviewerId, String decision, String notes) {
        this.reviewerId = reviewerId;
        this.decision = decision;
        this.notes = notes;
    }

    public String getReviewerId() { return reviewerId; }
    public void setReviewerId(String reviewerId) { this.reviewerId = reviewerId; }

    public String getDecision() { return decision; }
    public void setDecision(String decision) { this.decision = decision; }

    public String getNotes() { return notes; }
    public void setNotes(String notes) { this.notes = notes; }
}
