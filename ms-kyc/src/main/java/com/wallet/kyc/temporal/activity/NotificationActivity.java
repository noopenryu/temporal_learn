package com.wallet.kyc.temporal.activity;

import com.wallet.kyc.domain.KycStatus;
import io.temporal.activity.ActivityInterface;

@ActivityInterface
public interface NotificationActivity {
    void sendCustomerNotification(String customerId, String applicationId, KycStatus status, String message);
}
