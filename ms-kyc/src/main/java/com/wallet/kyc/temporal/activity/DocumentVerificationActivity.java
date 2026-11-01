package com.wallet.kyc.temporal.activity;

import com.wallet.kyc.temporal.model.DocumentVerificationResult;
import com.wallet.kyc.temporal.model.KycWorkflowRequest;
import io.temporal.activity.ActivityInterface;

@ActivityInterface
public interface DocumentVerificationActivity {
    DocumentVerificationResult verifyDocument(KycWorkflowRequest request);
}
