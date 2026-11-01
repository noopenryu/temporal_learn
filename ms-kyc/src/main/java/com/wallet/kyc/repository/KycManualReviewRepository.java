package com.wallet.kyc.repository;

import com.wallet.kyc.domain.entity.KycManualReviewEntity;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface KycManualReviewRepository extends JpaRepository<KycManualReviewEntity, String> {
    List<KycManualReviewEntity> findByApplicationIdOrderByReviewedAtDesc(String applicationId);
}
