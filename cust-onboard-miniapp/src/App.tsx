import React, { useState } from 'react';
import { Header } from './components/Header';
import { Stepper } from './components/Stepper';
import { StartSessionStep } from './components/steps/StartSessionStep';
import { PersonalInfoStep } from './components/steps/PersonalInfoStep';
import { AddressStep } from './components/steps/AddressStep';
import { IdentityDocsStep } from './components/steps/IdentityDocsStep';
import { FinancialStep } from './components/steps/FinancialStep';
import { WalletTierStep } from './components/steps/WalletTierStep';
import { ReviewStep } from './components/steps/ReviewStep';
import { LiveWorkflowTracker } from './components/workflow/LiveWorkflowTracker';
import {
  AddressInfo,
  ApplicationDetailResponse,
  FinancialInfo,
  IdentityDocuments,
  PersonalInfo,
  ResumeSessionRequest,
  StartSessionRequest,
  WalletTier,
} from './types/onboarding';
import { onboardingApi } from './api/onboardingApi';

export const App: React.FC = () => {
  const [detail, setDetail] = useState<ApplicationDetailResponse | null>(null);
  const [currentStep, setCurrentStep] = useState<number>(0);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [resumedBanner, setResumedBanner] = useState<string | null>(null);

  const application = detail?.application;
  const isDraft = !application || application.status === 'DRAFT';

  // Step map to recover last completed step when resuming
  const stepToNumber = (stepName?: string): number => {
    switch (stepName) {
      case 'personal_info':
        return 1;
      case 'address_info':
        return 2;
      case 'identity_documents':
        return 3;
      case 'financial_info':
        return 4;
      case 'wallet_tier':
        return 5;
      default:
        return 1;
    }
  };

  // Start a new onboarding session (or resume active if one exists for phone)
  const handleStartSession = async (req: StartSessionRequest) => {
    setIsLoading(true);
    setError(null);
    setResumedBanner(null);
    try {
      const res = await onboardingApi.startSession(req);
      setDetail(res);
      if (res.is_resumed) {
        setResumedBanner(`Resumed active application for ${res.application.personal_info?.phone_number || 'your phone number'}.`);
      }
      setCurrentStep(stepToNumber(res.application.current_step));
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  };

  // Resume existing application by Phone + Password
  const handleResumeSession = async (req: ResumeSessionRequest) => {
    setIsLoading(true);
    setError(null);
    setResumedBanner(null);
    try {
      const res = await onboardingApi.resumeSession(req);
      setDetail(res);
      setResumedBanner(`Resumed active application for ${res.application.personal_info?.phone_number || req.phone_number}.`);
      if (res.application.status === 'DRAFT') {
        setCurrentStep(stepToNumber(res.application.current_step));
      }
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  };

  // Step 1: Save Personal Info
  const handleSavePersonal = async (data: PersonalInfo) => {
    if (!application) return;
    setIsLoading(true);
    try {
      const res = await onboardingApi.patchPersonalInfo(application.id, data);
      setDetail(res);
      setCurrentStep(2);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  };

  // Step 2: Save Address Info
  const handleSaveAddress = async (data: AddressInfo) => {
    if (!application) return;
    setIsLoading(true);
    try {
      const res = await onboardingApi.patchAddress(application.id, data);
      setDetail(res);
      setCurrentStep(3);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  };

  // Step 3: Save Identity Docs
  const handleSaveIdentity = async (data: IdentityDocuments) => {
    if (!application) return;
    setIsLoading(true);
    try {
      const res = await onboardingApi.patchIdentityDocs(application.id, data);
      setDetail(res);
      setCurrentStep(4);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  };

  // Step 4: Save Financial Info
  const handleSaveFinancial = async (data: FinancialInfo) => {
    if (!application) return;
    setIsLoading(true);
    try {
      const res = await onboardingApi.patchFinancialInfo(application.id, data);
      setDetail(res);
      setCurrentStep(5);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  };

  // Step 5: Save Wallet Tier
  const handleSaveTier = async (tier: WalletTier) => {
    if (!application) return;
    setIsLoading(true);
    try {
      const res = await onboardingApi.patchWalletTier(application.id, tier);
      setDetail(res);
      setCurrentStep(6);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  };

  // Step 6: Submit Application
  const handleSubmitApplication = async () => {
    if (!application) return;
    setIsLoading(true);
    try {
      const res = await onboardingApi.submitApplication(application.id);
      setDetail(res);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  };

  // Cancel Application
  const handleCancelApplication = async () => {
    if (!application) return;
    const confirm = window.confirm('Are you sure you want to cancel this onboarding session?');
    if (!confirm) return;

    setIsLoading(true);
    try {
      await onboardingApi.cancelSession(application.id, 'User requested cancellation');
      setDetail(null);
      setCurrentStep(0);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  };

  // Reset to initial screen
  const handleReset = () => {
    setDetail(null);
    setCurrentStep(0);
    setError(null);
    setResumedBanner(null);
  };

  return (
    <div className="min-h-screen flex flex-col bg-slate-50 text-slate-900">
      <Header application={application} onReset={handleReset} />

      <main className="flex-1 max-w-6xl w-full mx-auto px-4 sm:px-6 py-6 sm:py-8">
        {/* If no active session, show Start/Resume */}
        {!detail && (
          <StartSessionStep
            onStart={handleStartSession}
            onResume={handleResumeSession}
            isLoading={isLoading}
            error={error}
            onClearError={() => setError(null)}
          />
        )}

        {/* If session exists and in DRAFT status, show Stepper + Current Form */}
        {detail && isDraft && (
          <div>
            {resumedBanner && (
              <div className="mb-6 p-3.5 rounded-xl bg-blue-50 border border-blue-200 text-blue-900 text-xs flex items-center justify-between">
                <span>{resumedBanner}</span>
                <button
                  type="button"
                  onClick={() => setResumedBanner(null)}
                  className="text-blue-700 hover:text-blue-900 font-bold ml-3 text-xs"
                >
                  Dismiss
                </button>
              </div>
            )}

            <Stepper
              currentStep={currentStep}
              completedSteps={detail.application.completed_steps || []}
              progressPercent={detail.progress_percent}
              onStepClick={(step) => setCurrentStep(step)}
            />

            {currentStep === 1 && (
              <PersonalInfoStep
                initialData={detail.application.personal_info}
                onSave={handleSavePersonal}
                onBack={() => setCurrentStep(0)}
                isLoading={isLoading}
              />
            )}

            {currentStep === 2 && (
              <AddressStep
                initialData={detail.application.address_info}
                onSave={handleSaveAddress}
                onBack={() => setCurrentStep(1)}
                isLoading={isLoading}
              />
            )}

            {currentStep === 3 && (
              <IdentityDocsStep
                initialData={detail.application.identity_docs}
                onSave={handleSaveIdentity}
                onBack={() => setCurrentStep(2)}
                isLoading={isLoading}
              />
            )}

            {currentStep === 4 && (
              <FinancialStep
                initialData={detail.application.financial_info}
                onSave={handleSaveFinancial}
                onBack={() => setCurrentStep(3)}
                isLoading={isLoading}
              />
            )}

            {currentStep === 5 && (
              <WalletTierStep
                initialTier={detail.application.wallet_tier_requested}
                onSave={handleSaveTier}
                onBack={() => setCurrentStep(4)}
                isLoading={isLoading}
              />
            )}

            {currentStep === 6 && (
              <ReviewStep
                detail={detail}
                onSubmit={handleSubmitApplication}
                onCancel={handleCancelApplication}
                onEditStep={(step) => setCurrentStep(step)}
                onBack={() => setCurrentStep(5)}
                isLoading={isLoading}
              />
            )}
          </div>
        )}

        {/* If session is submitted or past draft, render Live Workflow Tracker */}
        {detail && !isDraft && (
          <LiveWorkflowTracker initialDetail={detail} onNewSession={handleReset} />
        )}
      </main>

      <footer className="border-t border-slate-200/80 bg-white py-6 text-center text-xs text-slate-500">
        <p>PayFlow Internet Wallet • Powered by Temporal Saga Orchestration & Go Microservices</p>
      </footer>
    </div>
  );
};

export default App;
