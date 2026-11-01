package com.wallet.kyc.temporal.model;

import java.io.Serializable;

public class KycWorkflowProgress implements Serializable {
    private String currentStage;
    private int percentComplete;
    private String statusMessage;

    public KycWorkflowProgress() {}

    public KycWorkflowProgress(String currentStage, int percentComplete, String statusMessage) {
        this.currentStage = currentStage;
        this.percentComplete = percentComplete;
        this.statusMessage = statusMessage;
    }

    public String getCurrentStage() { return currentStage; }
    public void setCurrentStage(String currentStage) { this.currentStage = currentStage; }

    public int getPercentComplete() { return percentComplete; }
    public void setPercentComplete(int percentComplete) { this.percentComplete = percentComplete; }

    public String getStatusMessage() { return statusMessage; }
    public void setStatusMessage(String statusMessage) { this.statusMessage = statusMessage; }
}
