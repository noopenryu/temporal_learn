package com.wallet.kyc.dto;

public class WorkflowProgressResponseDto {

    private String applicationId;
    private String workflowId;
    private String currentStage;
    private int percentComplete;
    private String statusMessage;

    public WorkflowProgressResponseDto() {}

    public WorkflowProgressResponseDto(String applicationId, String workflowId, String currentStage, int percentComplete, String statusMessage) {
        this.applicationId = applicationId;
        this.workflowId = workflowId;
        this.currentStage = currentStage;
        this.percentComplete = percentComplete;
        this.statusMessage = statusMessage;
    }

    public String getApplicationId() { return applicationId; }
    public void setApplicationId(String applicationId) { this.applicationId = applicationId; }

    public String getWorkflowId() { return workflowId; }
    public void setWorkflowId(String workflowId) { this.workflowId = workflowId; }

    public String getCurrentStage() { return currentStage; }
    public void setCurrentStage(String currentStage) { this.currentStage = currentStage; }

    public int getPercentComplete() { return percentComplete; }
    public void setPercentComplete(int percentComplete) { this.percentComplete = percentComplete; }

    public String getStatusMessage() { return statusMessage; }
    public void setStatusMessage(String statusMessage) { this.statusMessage = statusMessage; }
}
