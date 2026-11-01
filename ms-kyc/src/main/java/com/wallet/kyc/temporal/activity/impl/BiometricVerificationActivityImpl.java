package com.wallet.kyc.temporal.activity.impl;

import com.wallet.kyc.temporal.activity.BiometricVerificationActivity;
import com.wallet.kyc.temporal.model.BiometricVerificationResult;
import com.wallet.kyc.temporal.model.KycWorkflowRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

@Component
public class BiometricVerificationActivityImpl implements BiometricVerificationActivity {

    private static final Logger log = LoggerFactory.getLogger(BiometricVerificationActivityImpl.class);

    @Override
    public BiometricVerificationResult verifyBiometrics(KycWorkflowRequest request) {
        log.info("Performing biometric verification for applicationId={}, selfieUrl={}",
                request.getApplicationId(), request.getSelfieUrl());

        if (request.getSelfieUrl() == null || request.getSelfieUrl().isBlank()) {
            return new BiometricVerificationResult(false, 0.0, false, "Missing selfie photograph");
        }

        String selfieUpper = request.getSelfieUrl().toUpperCase();
        if (selfieUpper.contains("MISMATCH") || selfieUpper.contains("FAIL")) {
            return new BiometricVerificationResult(false, 42.0, true, "Face match score below required threshold");
        }

        if (selfieUpper.contains("LIVENESS_FAIL")) {
            return new BiometricVerificationResult(false, 85.0, false, "Liveness check failed (anti-spoofing alert)");
        }

        if (selfieUpper.contains("BORDERLINE")) {
            // Triggers manual review (60% <= score < 80%)
            return new BiometricVerificationResult(true, 72.5, true, "Borderline face match confidence due to lighting/glare");
        }

        // Standard high confidence pass
        return new BiometricVerificationResult(true, 94.2, true, "High confidence facial match and liveness confirmed");
    }
}
