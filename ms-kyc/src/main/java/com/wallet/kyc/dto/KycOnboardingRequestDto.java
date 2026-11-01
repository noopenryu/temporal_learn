package com.wallet.kyc.dto;

import com.wallet.kyc.domain.DocumentType;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Past;

import java.time.LocalDate;

public class KycOnboardingRequestDto {

    @NotBlank(message = "customerId is required")
    private String customerId;

    @NotBlank(message = "fullName is required")
    private String fullName;

    @NotNull(message = "dateOfBirth is required")
    @Past(message = "dateOfBirth must be in the past")
    private LocalDate dateOfBirth;

    @NotBlank(message = "nationality is required")
    private String nationality;

    @NotBlank(message = "phoneNumber is required")
    private String phoneNumber;

    @NotBlank(message = "email is required")
    @Email(message = "email must be valid")
    private String email;

    @NotNull(message = "documentType is required")
    private DocumentType documentType;

    @NotBlank(message = "documentNumber is required")
    private String documentNumber;

    @NotBlank(message = "frontDocumentUrl is required")
    private String frontDocumentUrl;

    private String backDocumentUrl;

    @NotBlank(message = "selfieUrl is required")
    private String selfieUrl;

    public KycOnboardingRequestDto() {}

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
