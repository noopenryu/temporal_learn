package com.wallet.kyc.repository;

import com.wallet.kyc.domain.entity.KycVerificationRecordEntity;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface KycVerificationRecordRepository extends JpaRepository<KycVerificationRecordEntity, String> {
    List<KycVerificationRecordEntity> findByApplicationIdOrderByCreatedAtAsc(String applicationId);
}
