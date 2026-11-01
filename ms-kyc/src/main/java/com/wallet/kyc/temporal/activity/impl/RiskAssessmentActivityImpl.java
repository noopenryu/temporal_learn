package com.wallet.kyc.temporal.activity.impl;

import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.repository.KycApplicationRepository;
import com.wallet.kyc.temporal.activity.RiskAssessmentActivity;
import com.wallet.kyc.temporal.model.BiometricVerificationResult;
import com.wallet.kyc.temporal.model.DocumentVerificationResult;
import com.wallet.kyc.temporal.model.KycWorkflowRequest;
import com.wallet.kyc.temporal.model.RiskAssessmentResult;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.time.LocalDate;
import java.time.Period;

@Component
public class RiskAssessmentActivityImpl implements RiskAssessmentActivity {

    private static final Logger log = LoggerFactory.getLogger(RiskAssessmentActivityImpl.class);

    private final KycApplicationRepository applicationRepository;

    public RiskAssessmentActivityImpl(KycApplicationRepository applicationRepository) {
        this.applicationRepository = applicationRepository;
    }

    @Override
    public RiskAssessmentResult assessRisk(KycWorkflowRequest request,
                                           DocumentVerificationResult docResult,
                                           BiometricVerificationResult bioResult) {
        log.info("Assessing risk for customerId={}, dob={}", request.getCustomerId(), request.getDateOfBirth());

        // 1. Age Eligibility Check (Minimum 18 years old)
        int age = Period.between(request.getDateOfBirth(), LocalDate.now()).getYears();
        boolean ageEligible = age >= 18;

        // 2. Duplicate Detection
        boolean duplicateDetected = false;
        try {
            duplicateDetected = applicationRepository.existsByDocumentTypeAndDocumentNumberAndStatus(
                    request.getDocumentType(), request.getDocumentNumber(), KycStatus.APPROVED);
        } catch (Exception e) {
            log.warn("Could not check duplicate document status in DB (using false): {}", e.getMessage());
        }

        // 3. Risk Score Calculation (0 = minimal risk, 100 = extreme risk)
        int score = 15; // baseline risk
        if (age < 21) {
            score += 15;
        }
        if (bioResult.getFaceMatchScore() < 80.0) {
            score += 25;
        }
        if (request.getNationality() == null || request.getNationality().length() > 3) {
            score += 10;
        }

        String riskLevel = (score < 30) ? "LOW" : (score < 60 ? "MEDIUM" : "HIGH");
        String summary = String.format("Applicant age %d (eligible=%b), duplicate=%b, composite score=%d (%s)",
                age, ageEligible, duplicateDetected, score, riskLevel);

        return new RiskAssessmentResult(score, ageEligible, duplicateDetected, riskLevel, summary);
    }
}
