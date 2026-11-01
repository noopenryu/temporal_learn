import React, { useState, useEffect, useCallback } from 'react';
import { Header } from './components/Header';
import { StatsBar } from './components/StatsBar';
import { QueueFilters } from './components/QueueFilters';
import { OnboardingQueueTable } from './components/OnboardingQueueTable';
import { KycQueueTable } from './components/KycQueueTable';
import { CaseReviewModal } from './components/CaseReviewModal';
import { KycReviewModal } from './components/KycReviewModal';
import { staffApi } from './api/staffApi';
import { ApplicationDetailResponse, KycApplicationDto } from './types/staff';
import { CheckCircle2, AlertCircle } from 'lucide-react';

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'onboarding' | 'kyc'>('onboarding');
  const [statusFilter, setStatusFilter] = useState<string>('REQUIRES_MANUAL_REVIEW');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [reviewerId, setReviewerId] = useState<string>('OFFICER-COMPLIANCE-01');

  const [onboardingItems, setOnboardingItems] = useState<ApplicationDetailResponse[]>([]);
  const [kycItems, setKycItems] = useState<KycApplicationDto[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);
  const [isSeeding, setIsSeeding] = useState<boolean>(false);
  const [autoRefresh, setAutoRefresh] = useState<boolean>(true);

  const [selectedOnboarding, setSelectedOnboarding] = useState<ApplicationDetailResponse | null>(null);
  const [selectedKyc, setSelectedKyc] = useState<KycApplicationDto | null>(null);

  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  const showToast = (message: string, type: 'success' | 'error' = 'success') => {
    setToast({ message, type });
    setTimeout(() => {
      setToast(null);
    }, 4500);
  };

  // Fetch Onboarding applications
  const fetchOnboarding = useCallback(async (quiet = false) => {
    if (!quiet) setIsLoading(true);
    try {
      const res = await staffApi.listApplications({
        status: statusFilter,
        search: searchQuery,
      });
      setOnboardingItems(res.items || []);
    } catch (err) {
      console.error('Failed to fetch onboarding applications:', err);
    } finally {
      if (!quiet) setIsLoading(false);
    }
  }, [statusFilter, searchQuery]);

  // Fetch KYC applications
  const fetchKyc = useCallback(async (quiet = false) => {
    if (!quiet) setIsLoading(true);
    try {
      const res = await staffApi.listKycApplications(statusFilter);
      setKycItems(res || []);
    } catch (err) {
      console.error('Failed to fetch KYC applications:', err);
    } finally {
      if (!quiet) setIsLoading(false);
    }
  }, [statusFilter]);

  // Load active tab data
  const loadData = useCallback(async (quiet = false) => {
    if (activeTab === 'onboarding') {
      await fetchOnboarding(quiet);
    } else {
      await fetchKyc(quiet);
    }
  }, [activeTab, fetchOnboarding, fetchKyc]);

  // On mount and when dependencies change
  useEffect(() => {
    loadData();
  }, [loadData]);

  // Auto-refresh interval (10 seconds)
  useEffect(() => {
    if (!autoRefresh) return;
    const interval = setInterval(() => {
      loadData(true);
    }, 10000);
    return () => clearInterval(interval);
  }, [autoRefresh, loadData]);

  // Manual refresh trigger
  const handleManualRefresh = async () => {
    setIsRefreshing(true);
    await loadData(false);
    setIsRefreshing(false);
    showToast('Queue data synced successfully.');
  };

  // Seed demo case
  const handleSeedDemo = async (scenario: 'review_name' | 'pep' | 'standard') => {
    setIsSeeding(true);
    try {
      if (activeTab === 'onboarding') {
        const res = await staffApi.seedDemoApplication(scenario);
        showToast(
          `Demo customer application created for ${res.application.personal_info?.first_name} ${res.application.personal_info?.last_name}! Temporal workflow started.`
        );
        await fetchOnboarding();
        // If in manual review state or requires review, auto-select it for convenient testing
        if (res.application.status === 'REQUIRES_MANUAL_REVIEW') {
          setSelectedOnboarding(res);
        }
      } else {
        const kycRes = await staffApi.seedDemoKyc(scenario === 'review_name' ? 'borderline' : 'clear');
        showToast(`Demo KYC case created for ${kycRes.fullName}!`);
        await fetchKyc();
      }
    } catch (err) {
      showToast((err as Error).message, 'error');
    } finally {
      setIsSeeding(false);
    }
  };

  // Filtered items (client-side search fallback if needed)
  const filteredOnboarding = onboardingItems.filter((item) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    const app = item.application;
    const name = `${app.personal_info?.first_name || ''} ${app.personal_info?.last_name || ''}`.toLowerCase();
    const phone = (app.personal_info?.phone_number || '').toLowerCase();
    const email = (app.personal_info?.email || '').toLowerCase();
    const custId = app.customer_id.toLowerCase();
    const id = app.id.toLowerCase();
    return name.includes(q) || phone.includes(q) || email.includes(q) || custId.includes(q) || id.includes(q);
  });

  const filteredKyc = kycItems.filter((item) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      item.fullName?.toLowerCase().includes(q) ||
      item.customerId?.toLowerCase().includes(q) ||
      item.applicationId?.toLowerCase().includes(q) ||
      item.documentNumber?.toLowerCase().includes(q)
    );
  });

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-sans">
      {/* App Header */}
      <Header
        reviewerId={reviewerId}
        onReviewerIdChange={setReviewerId}
        onRefresh={handleManualRefresh}
        isRefreshing={isRefreshing}
      />

      {/* Main Content Dashboard */}
      <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6 space-y-6">
        {/* KPI Stats Bar */}
        <StatsBar
          items={onboardingItems}
          onSelectFilter={(st) => setStatusFilter(st)}
          activeStatus={statusFilter}
        />

        {/* Queue Switcher, Filter Chips & Search Bar */}
        <QueueFilters
          activeTab={activeTab}
          onTabChange={setActiveTab}
          statusFilter={statusFilter}
          onStatusFilterChange={setStatusFilter}
          searchQuery={searchQuery}
          onSearchQueryChange={setSearchQuery}
          onSeedDemo={handleSeedDemo}
          isSeeding={isSeeding}
          autoRefresh={autoRefresh}
          onAutoRefreshToggle={() => setAutoRefresh(!autoRefresh)}
        />

        {/* Data Queue Tables */}
        {activeTab === 'onboarding' ? (
          <OnboardingQueueTable
            items={filteredOnboarding}
            isLoading={isLoading}
            onSelectApplication={(item) => setSelectedOnboarding(item)}
            onSeedFirst={() => handleSeedDemo('review_name')}
          />
        ) : (
          <KycQueueTable
            items={filteredKyc}
            isLoading={isLoading}
            onSelectKyc={(kyc) => setSelectedKyc(kyc)}
            onSeedKyc={() => handleSeedDemo('review_name')}
          />
        )}
      </main>

      {/* Case Review Modal for Customer Onboarding */}
      {selectedOnboarding && (
        <CaseReviewModal
          detail={selectedOnboarding}
          reviewerId={reviewerId}
          onClose={() => setSelectedOnboarding(null)}
          onSuccess={(msg) => {
            showToast(msg);
            fetchOnboarding(true);
          }}
        />
      )}

      {/* KYC Review Modal */}
      {selectedKyc && (
        <KycReviewModal
          kyc={selectedKyc}
          reviewerId={reviewerId}
          onClose={() => setSelectedKyc(null)}
          onSuccess={(msg) => {
            showToast(msg);
            fetchKyc(true);
          }}
        />
      )}

      {/* Toast Notification */}
      {toast && (
        <div className="fixed bottom-5 right-5 z-50 flex items-center space-x-2 px-4 py-3 rounded-xl bg-slate-900 border border-slate-700 text-xs shadow-2xl animate-in slide-in-from-bottom-5">
          {toast.type === 'success' ? (
            <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
          ) : (
            <AlertCircle className="w-4 h-4 text-rose-400 shrink-0" />
          )}
          <span className="font-medium text-white">{toast.message}</span>
        </div>
      )}
    </div>
  );
};
