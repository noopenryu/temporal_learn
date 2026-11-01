package com.wallet.kyc.temporal.model;

import java.io.Serializable;

public class SanctionsScreeningResult implements Serializable {
    private boolean clear;
    private boolean pepMatch;
    private boolean sanctionsMatch;
    private String matchedList;

    public SanctionsScreeningResult() {}

    public SanctionsScreeningResult(boolean clear, boolean pepMatch, boolean sanctionsMatch, String matchedList) {
        this.clear = clear;
        this.pepMatch = pepMatch;
        this.sanctionsMatch = sanctionsMatch;
        this.matchedList = matchedList;
    }

    public boolean isClear() { return clear; }
    public void setClear(boolean clear) { this.clear = clear; }

    public boolean isPepMatch() { return pepMatch; }
    public void setPepMatch(boolean pepMatch) { this.pepMatch = pepMatch; }

    public boolean isSanctionsMatch() { return sanctionsMatch; }
    public void setSanctionsMatch(boolean sanctionsMatch) { this.sanctionsMatch = sanctionsMatch; }

    public String getMatchedList() { return matchedList; }
    public void setMatchedList(String matchedList) { this.matchedList = matchedList; }
}
