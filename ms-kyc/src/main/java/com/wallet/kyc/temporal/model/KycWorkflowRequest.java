package com.wallet.kyc.temporal.model;

import com.wallet.kyc.domain.DocumentType;
import java.io.Serializable;
import java.time.LocalDate;

public class KycWorkflowRequest implements Serializable {
    private String applicationId;
    private String customerId;
    private String fullName;
    private LocalDate dateOfBirth;
    private String nationality;
    private String phoneNumber;
    private String email;
    private DocumentType documentType;
    private String documentNumber;
    private String frontDocumentUrl;
    private String backDocumentUrl;
    private String selfieUrl;

    public KycWorkflowRequest() {}

    public KycWorkflowRequest(String applicationId, String customerId, String fullName, LocalDate dateOfBirth,
                              String nationality, String phoneNumber, String email, DocumentType documentType,
                              String documentNumber, String frontDocumentUrl, String backDocumentUrl, String selfieUrl) {
        this.applicationId = applicationId;
        this.customerId = customerId;
        this.fullName = fullName;
        this.dateOfBirth = dateOfBirth;
        this.nationality = nationality;
        this.phoneNumber = phoneNumber;
        this.email = email;
        this.documentType = documentType;
        this.documentNumber = documentNumber;
        this.frontDocumentUrl = frontDocumentUrl;
        this.backDocumentUrl = backDocumentUrl;
        this.selfieUrl = selfieUrl;
    }

    public String getApplicationId() { return applicationId; }
    public void setApplicationId(String applicationId) { this.applicationId = applicationId; }

    public String getCustomerId() { return customerId; }
    public void setCustomerId(String customerId) { this.customerId = customerId; }

    public String getFullName() { return fullName; }
    public void setFullName(String fullName) { this.fullName = fullName; }

    public LocalDate getDateOfBirth() { return dateOfBirth; }
    public void setDateOfBirth(LocalDate dateOfBirth) { this.dateOfBirth = dateOfBirth; }

    public String getNationality() { return nationality; }
    public void setNationality(String nationality) { this.nationality = nationality; }

    public String getPhoneNumber() { return phoneNumber; }
    public void setPhoneNumber(String phoneNumber) { this.phoneNumber = phoneNumber; }

    public String getEmail() { return email; }
    public void setEmail(String email) { this.email = email; }

    public DocumentType getDocumentType() { return documentType; }
    public void setDocumentType(DocumentType documentType) { this.documentType = documentType; }

    public String getDocumentNumber() { return documentNumber; }
    public void setDocumentNumber(String documentNumber) { this.documentNumber = documentNumber; }

    public String getFrontDocumentUrl() { return frontDocumentUrl; }
    public void setFrontDocumentUrl(String frontDocumentUrl) { this.frontDocumentUrl = frontDocumentUrl; }

    public String getBackDocumentUrl() { return backDocumentUrl; }
    public void setBackDocumentUrl(String backDocumentUrl) { this.backDocumentUrl = backDocumentUrl; }

    public String getSelfieUrl() { return selfieUrl; }
    public void setSelfieUrl(String selfieUrl) { this.selfieUrl = selfieUrl; }
}
