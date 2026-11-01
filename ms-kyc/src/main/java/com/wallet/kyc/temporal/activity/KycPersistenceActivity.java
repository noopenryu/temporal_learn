package com.wallet.kyc.temporal.activity;

import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.domain.WalletTier;
import io.temporal.activity.ActivityInterface;

@ActivityInterface
public interface KycPersistenceActivity {
    void recordActivityAudit(String applicationId, String stepName, boolean passed, Double score, String detailsJson);

    void updateApplicationStatus(String applicationId, KycStatus status, WalletTier tier, Integer riskScore, String rejectionReason);

    void recordManualReview(String applicationId, String reviewerId, String decision, String notes);
}
