package com.wallet.kyc.temporal.model;

import java.io.Serializable;

public class BiometricVerificationResult implements Serializable {
    private boolean matched;
    private double faceMatchScore;
    private boolean livenessPassed;
    private String details;

    public BiometricVerificationResult() {}

    public BiometricVerificationResult(boolean matched, double faceMatchScore, boolean livenessPassed, String details) {
        this.matched = matched;
        this.faceMatchScore = faceMatchScore;
        this.livenessPassed = livenessPassed;
        this.details = details;
    }

    public boolean isMatched() { return matched; }
    public void setMatched(boolean matched) { this.matched = matched; }

    public double getFaceMatchScore() { return faceMatchScore; }
    public void setFaceMatchScore(double faceMatchScore) { this.faceMatchScore = faceMatchScore; }

    public boolean isLivenessPassed() { return livenessPassed; }
    public void setLivenessPassed(boolean livenessPassed) { this.livenessPassed = livenessPassed; }

    public String getDetails() { return details; }
    public void setDetails(String details) { this.details = details; }
}
