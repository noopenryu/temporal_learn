package com.wallet.kyc.temporal.activity;

import com.wallet.kyc.temporal.model.BiometricVerificationResult;
import com.wallet.kyc.temporal.model.KycWorkflowRequest;
import io.temporal.activity.ActivityInterface;

@ActivityInterface
public interface BiometricVerificationActivity {
    BiometricVerificationResult verifyBiometrics(KycWorkflowRequest request);
}
