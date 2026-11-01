package com.wallet.kyc.temporal.model;

import java.io.Serializable;

public class RiskAssessmentResult implements Serializable {
    private int riskScore;
    private boolean ageEligible;
    private boolean duplicateDetected;
    private String riskLevel;
    private String summary;

    public RiskAssessmentResult() {}

    public RiskAssessmentResult(int riskScore, boolean ageEligible, boolean duplicateDetected, String riskLevel, String summary) {
        this.riskScore = riskScore;
        this.ageEligible = ageEligible;
        this.duplicateDetected = duplicateDetected;
        this.riskLevel = riskLevel;
        this.summary = summary;
    }

    public int getRiskScore() { return riskScore; }
    public void setRiskScore(int riskScore) { this.riskScore = riskScore; }

    public boolean isAgeEligible() { return ageEligible; }
    public void setAgeEligible(boolean ageEligible) { this.ageEligible = ageEligible; }

    public boolean isDuplicateDetected() { return duplicateDetected; }
    public void setDuplicateDetected(boolean duplicateDetected) { this.duplicateDetected = duplicateDetected; }

    public String getRiskLevel() { return riskLevel; }
    public void setRiskLevel(String riskLevel) { this.riskLevel = riskLevel; }

    public String getSummary() { return summary; }
    public void setSummary(String summary) { this.summary = summary; }
}
