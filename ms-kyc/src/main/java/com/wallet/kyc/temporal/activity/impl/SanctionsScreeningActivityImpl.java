package com.wallet.kyc.temporal.activity.impl;

import com.wallet.kyc.temporal.activity.SanctionsScreeningActivity;
import com.wallet.kyc.temporal.model.KycWorkflowRequest;
import com.wallet.kyc.temporal.model.SanctionsScreeningResult;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.util.Set;

@Component
public class SanctionsScreeningActivityImpl implements SanctionsScreeningActivity {

    private static final Logger log = LoggerFactory.getLogger(SanctionsScreeningActivityImpl.class);

    private static final Set<String> HIGH_RISK_COUNTRIES = Set.of("NK", "IR", "SY");

    @Override
    public SanctionsScreeningResult screenSanctions(KycWorkflowRequest request) {
        log.info("Screening sanctions and PEP for applicant={}, nationality={}",
                request.getFullName(), request.getNationality());

        String nameUpper = request.getFullName().toUpperCase();

        if (nameUpper.contains("SANCTION") || nameUpper.contains("TERROR") || nameUpper.contains("BLOCKED")) {
            return new SanctionsScreeningResult(false, false, true, "OFAC / UN Global Sanctions List");
        }

        if (nameUpper.contains("PEP") || nameUpper.contains("POLITICIAN")) {
            return new SanctionsScreeningResult(false, true, false, "PEP (Politically Exposed Persons) Watchlist");
        }

        if (request.getNationality() != null && HIGH_RISK_COUNTRIES.contains(request.getNationality().toUpperCase())) {
            return new SanctionsScreeningResult(false, false, true, "FATF High Risk Jurisdiction Blacklist (" + request.getNationality() + ")");
        }

        return new SanctionsScreeningResult(true, false, false, "None (Clean)");
    }
}
