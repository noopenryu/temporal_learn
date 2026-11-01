package com.wallet.kyc.temporal.activity;

import com.wallet.kyc.temporal.model.BiometricVerificationResult;
import com.wallet.kyc.temporal.model.DocumentVerificationResult;
import com.wallet.kyc.temporal.model.KycWorkflowRequest;
import com.wallet.kyc.temporal.model.RiskAssessmentResult;
import io.temporal.activity.ActivityInterface;

@ActivityInterface
public interface RiskAssessmentActivity {
    RiskAssessmentResult assessRisk(KycWorkflowRequest request,
                                    DocumentVerificationResult docResult,
                                    BiometricVerificationResult bioResult);
}
