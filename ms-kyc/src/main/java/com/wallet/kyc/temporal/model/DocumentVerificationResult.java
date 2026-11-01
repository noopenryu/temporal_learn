package com.wallet.kyc.temporal.model;

import java.io.Serializable;

public class DocumentVerificationResult implements Serializable {
    private boolean valid;
    private String failureReason;
    private double confidenceScore;
    private String extractedDocumentNumber;

    public DocumentVerificationResult() {}

    public DocumentVerificationResult(boolean valid, String failureReason, double confidenceScore, String extractedDocumentNumber) {
        this.valid = valid;
        this.failureReason = failureReason;
        this.confidenceScore = confidenceScore;
        this.extractedDocumentNumber = extractedDocumentNumber;
    }

    public boolean isValid() { return valid; }
    public void setValid(boolean valid) { this.valid = valid; }

    public String getFailureReason() { return failureReason; }
    public void setFailureReason(String failureReason) { this.failureReason = failureReason; }

    public double getConfidenceScore() { return confidenceScore; }
    public void setConfidenceScore(double confidenceScore) { this.confidenceScore = confidenceScore; }

    public String getExtractedDocumentNumber() { return extractedDocumentNumber; }
    public void setExtractedDocumentNumber(String extractedDocumentNumber) { this.extractedDocumentNumber = extractedDocumentNumber; }
}
