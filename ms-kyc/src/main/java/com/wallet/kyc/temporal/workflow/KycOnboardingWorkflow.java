package com.wallet.kyc.temporal.workflow;

import com.wallet.kyc.temporal.model.KycWorkflowProgress;
import com.wallet.kyc.temporal.model.KycWorkflowRequest;
import com.wallet.kyc.temporal.model.KycWorkflowResult;
import com.wallet.kyc.temporal.model.ManualReviewDecision;
import io.temporal.workflow.QueryMethod;
import io.temporal.workflow.SignalMethod;
import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;

@WorkflowInterface
public interface KycOnboardingWorkflow {

    @WorkflowMethod
    KycWorkflowResult executeKyc(KycWorkflowRequest request);

    @SignalMethod
    void submitManualReview(ManualReviewDecision decision);

    @QueryMethod
    String getWorkflowStatus();

    @QueryMethod
    KycWorkflowProgress getProgress();
}
