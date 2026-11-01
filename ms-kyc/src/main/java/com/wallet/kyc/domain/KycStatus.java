package com.wallet.kyc.domain;

public enum KycStatus {
    PENDING,
    PROCESSING,
    REQUIRES_MANUAL_REVIEW,
    APPROVED,
    REJECTED
}
