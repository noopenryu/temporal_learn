package com.wallet.kyc.temporal.activity.impl;

import com.wallet.kyc.temporal.activity.DocumentVerificationActivity;
import com.wallet.kyc.temporal.model.DocumentVerificationResult;
import com.wallet.kyc.temporal.model.KycWorkflowRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

@Component
public class DocumentVerificationActivityImpl implements DocumentVerificationActivity {

    private static final Logger log = LoggerFactory.getLogger(DocumentVerificationActivityImpl.class);

    @Override
    public DocumentVerificationResult verifyDocument(KycWorkflowRequest request) {
        log.info("Verifying document for applicationId={}, docType={}, docNumber={}",
                request.getApplicationId(), request.getDocumentType(), request.getDocumentNumber());

        if (request.getDocumentNumber() == null || request.getDocumentNumber().trim().length() < 5) {
            return new DocumentVerificationResult(false, "Invalid document number format", 0.0, null);
        }

        String docUpper = request.getDocumentNumber().toUpperCase();
        if (docUpper.contains("EXPIRED")) {
            return new DocumentVerificationResult(false, "Document has expired", 20.0, request.getDocumentNumber());
        }

        if (docUpper.contains("FORGED") || docUpper.contains("INVALID")) {
            return new DocumentVerificationResult(false, "Document authenticity verification failed (possible forgery)", 10.0, request.getDocumentNumber());
        }

        if (request.getFrontDocumentUrl() == null || request.getFrontDocumentUrl().isBlank()) {
            return new DocumentVerificationResult(false, "Missing front document image", 0.0, null);
        }

        // Successfully verified
        return new DocumentVerificationResult(true, null, 98.5, request.getDocumentNumber());
    }
}
