package com.wallet.kyc.temporal.activity;

import com.wallet.kyc.temporal.model.KycWorkflowRequest;
import com.wallet.kyc.temporal.model.SanctionsScreeningResult;
import io.temporal.activity.ActivityInterface;

@ActivityInterface
public interface SanctionsScreeningActivity {
    SanctionsScreeningResult screenSanctions(KycWorkflowRequest request);
}
