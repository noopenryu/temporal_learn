package com.wallet.kyc.temporal.activity.impl;

import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.temporal.activity.NotificationActivity;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

@Component
public class NotificationActivityImpl implements NotificationActivity {

    private static final Logger log = LoggerFactory.getLogger(NotificationActivityImpl.class);

    @Override
    public void sendCustomerNotification(String customerId, String applicationId, KycStatus status, String message) {
        log.info("DISPATCH NOTIFICATION -> customerId={}, applicationId={}, status={}: '{}'",
                customerId, applicationId, status, message);
    }
}
