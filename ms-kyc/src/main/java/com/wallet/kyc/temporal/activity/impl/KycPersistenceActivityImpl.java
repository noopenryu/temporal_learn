package com.wallet.kyc.temporal.activity.impl;

import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.domain.WalletTier;
import com.wallet.kyc.domain.entity.KycApplicationEntity;
import com.wallet.kyc.domain.entity.KycManualReviewEntity;
import com.wallet.kyc.domain.entity.KycVerificationRecordEntity;
import com.wallet.kyc.repository.KycApplicationRepository;
import com.wallet.kyc.repository.KycManualReviewRepository;
import com.wallet.kyc.repository.KycVerificationRecordRepository;
import com.wallet.kyc.temporal.activity.KycPersistenceActivity;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Component
public class KycPersistenceActivityImpl implements KycPersistenceActivity {

    private static final Logger log = LoggerFactory.getLogger(KycPersistenceActivityImpl.class);

    private final KycApplicationRepository applicationRepository;
    private final KycVerificationRecordRepository recordRepository;
    private final KycManualReviewRepository manualReviewRepository;

    public KycPersistenceActivityImpl(KycApplicationRepository applicationRepository,
                                      KycVerificationRecordRepository recordRepository,
                                      KycManualReviewRepository manualReviewRepository) {
        this.applicationRepository = applicationRepository;
        this.recordRepository = recordRepository;
        this.manualReviewRepository = manualReviewRepository;
    }

    @Override
    @Transactional
    public void recordActivityAudit(String applicationId, String stepName, boolean passed, Double score, String detailsJson) {
        log.info("Persisting audit record for applicationId={}, step={}, passed={}", applicationId, stepName, passed);
        String recordId = UUID.randomUUID().toString();
        KycVerificationRecordEntity record = new KycVerificationRecordEntity(
                recordId, applicationId, stepName, passed, score, detailsJson
        );
        recordRepository.save(record);
    }

    @Override
    @Transactional
    public void updateApplicationStatus(String applicationId, KycStatus status, WalletTier tier, Integer riskScore, String rejectionReason) {
        log.info("Updating application status for applicationId={}, status={}, tier={}, riskScore={}",
                applicationId, status, tier, riskScore);

        applicationRepository.findById(applicationId).ifPresent(entity -> {
            entity.setStatus(status);
            if (tier != null) {
                entity.setWalletTier(tier);
            }
            if (riskScore != null) {
                entity.setRiskScore(riskScore);
            }
            if (rejectionReason != null) {
                entity.setRejectionReason(rejectionReason);
            }
            applicationRepository.save(entity);
        });
    }

    @Override
    @Transactional
    public void recordManualReview(String applicationId, String reviewerId, String decision, String notes) {
        log.info("Recording manual review for applicationId={}, reviewer={}, decision={}",
                applicationId, reviewerId, decision);

        String reviewId = UUID.randomUUID().toString();
        KycManualReviewEntity entity = new KycManualReviewEntity(
                reviewId, applicationId, reviewerId, decision, notes
        );
        manualReviewRepository.save(entity);
    }
}
