package com.wallet.kyc.repository;

import com.wallet.kyc.domain.DocumentType;
import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.domain.entity.KycApplicationEntity;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface KycApplicationRepository extends JpaRepository<KycApplicationEntity, String> {
    List<KycApplicationEntity> findByCustomerIdOrderByCreatedAtDesc(String customerId);
    Optional<KycApplicationEntity> findFirstByCustomerIdOrderByCreatedAtDesc(String customerId);
    Optional<KycApplicationEntity> findByWorkflowId(String workflowId);
    List<KycApplicationEntity> findByStatusOrderByCreatedAtDesc(KycStatus status);
    boolean existsByDocumentTypeAndDocumentNumberAndStatus(DocumentType documentType, String documentNumber, KycStatus status);
}
