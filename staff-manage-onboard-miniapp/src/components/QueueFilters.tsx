import React, { useState } from 'react';
import { Search, Sparkles, ChevronDown } from 'lucide-react';

interface QueueFiltersProps {
  activeTab: 'onboarding' | 'kyc';
  onTabChange: (tab: 'onboarding' | 'kyc') => void;
  statusFilter: string;
  onStatusFilterChange: (status: string) => void;
  searchQuery: string;
  onSearchQueryChange: (q: string) => void;
  onSeedDemo: (scenario: 'review_name' | 'pep' | 'standard') => Promise<void>;
  isSeeding: boolean;
  autoRefresh: boolean;
  onAutoRefreshToggle: () => void;
}

export const QueueFilters: React.FC<QueueFiltersProps> = ({
  activeTab,
  onTabChange,
  statusFilter,
  onStatusFilterChange,
  searchQuery,
  onSearchQueryChange,
  onSeedDemo,
  isSeeding,
  autoRefresh,
  onAutoRefreshToggle,
}) => {
  const [showSeedDropdown, setShowSeedDropdown] = useState(false);

  const statusChips = [
    { id: 'ALL', label: 'All Cases' },
    { id: 'REQUIRES_MANUAL_REVIEW', label: 'Requires Review' },
    { id: 'PROCESSING', label: 'In-Flight' },
    { id: 'WAITING_UPDATED_DOCS', label: 'Waiting Docs' },
    { id: 'APPROVED', label: 'Approved' },
    { id: 'REJECTED', label: 'Rejected' },
    { id: 'DRAFT', label: 'Draft' },
  ];

  return (
    <div className="space-y-3.5">
      {/* Top row: Queue tab selector + Seed button + Auto-refresh */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        {/* Tab selection */}
        <div className="flex items-center bg-slate-900 border border-slate-800 p-1 rounded-xl">
          <button
            type="button"
            onClick={() => onTabChange('onboarding')}
            className={`px-3.5 py-1.5 rounded-lg text-xs font-semibold transition ${
              activeTab === 'onboarding'
                ? 'bg-slate-800 text-cyan-400 border border-slate-700 shadow-xs'
                : 'text-slate-400 hover:text-white'
            }`}
          >
            Customer Onboarding Queue
          </button>
          <button
            type="button"
            onClick={() => onTabChange('kyc')}
            className={`px-3.5 py-1.5 rounded-lg text-xs font-semibold transition ${
              activeTab === 'kyc'
                ? 'bg-slate-800 text-cyan-400 border border-slate-700 shadow-xs'
                : 'text-slate-400 hover:text-white'
            }`}
          >
            KYC Microservice Cases
          </button>
        </div>

        {/* Right tools: Auto-refresh + Seed Case button */}
        <div className="flex items-center space-x-3 w-full sm:w-auto justify-between sm:justify-end">
          <label className="flex items-center space-x-2 text-xs text-slate-400 cursor-pointer select-none">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={onAutoRefreshToggle}
              className="rounded bg-slate-800 border-slate-700 text-cyan-500 focus:ring-0 focus:ring-offset-0"
            />
            <span className="font-mono text-[11px]">Auto-poll (10s)</span>
          </label>

          {/* Seed demo cases button */}
          <div className="relative">
            <button
              type="button"
              disabled={isSeeding}
              onClick={() => setShowSeedDropdown(!showSeedDropdown)}
              className="px-3 py-1.5 rounded-lg bg-amber-500/15 border border-amber-500/30 text-amber-300 hover:bg-amber-500/25 text-xs font-semibold flex items-center space-x-1.5 transition disabled:opacity-50"
            >
              <Sparkles className="w-3.5 h-3.5 text-amber-400" />
              <span>{isSeeding ? 'Generating...' : 'Seed Test Case'}</span>
              <ChevronDown className="w-3.5 h-3.5 ml-1 opacity-70" />
            </button>

            {showSeedDropdown && (
              <div className="absolute right-0 mt-2 w-72 bg-slate-900 border border-slate-800 rounded-xl shadow-2xl p-1.5 z-40 space-y-1">
                <button
                  type="button"
                  onClick={() => {
                    setShowSeedDropdown(false);
                    onSeedDemo('review_name');
                  }}
                  className="w-full text-left p-2.5 rounded-lg hover:bg-slate-800 text-xs transition"
                >
                  <p className="font-bold text-white">Review Flagged Case</p>
                  <p className="text-[11px] text-slate-400">Triggers manual review for name mismatch or borderline OCR</p>
                </button>

                <button
                  type="button"
                  onClick={() => {
                    setShowSeedDropdown(false);
                    onSeedDemo('pep');
                  }}
                  className="w-full text-left p-2.5 rounded-lg hover:bg-slate-800 text-xs transition"
                >
                  <p className="font-bold text-white">Politically Exposed Person (PEP)</p>
                  <p className="text-[11px] text-slate-400">Declared PEP requiring enhanced due diligence approval</p>
                </button>

                <button
                  type="button"
                  onClick={() => {
                    setShowSeedDropdown(false);
                    onSeedDemo('standard');
                  }}
                  className="w-full text-left p-2.5 rounded-lg hover:bg-slate-800 text-xs transition"
                >
                  <p className="font-bold text-white">Standard Pass Case</p>
                  <p className="text-[11px] text-slate-400">Clean applicant automatically approved by Temporal saga</p>
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Bottom row: Status Filter Chips + Search Bar */}
      <div className="flex flex-col md:flex-row items-stretch md:items-center justify-between gap-3">
        {/* Chips */}
        <div className="flex items-center space-x-1.5 overflow-x-auto pb-1 md:pb-0 scrollbar-none">
          {statusChips.map((chip) => {
            const isActive = statusFilter === chip.id;
            return (
              <button
                key={chip.id}
                type="button"
                onClick={() => onStatusFilterChange(chip.id)}
                className={`px-3 py-1.5 rounded-lg text-xs font-semibold whitespace-nowrap transition border ${
                  isActive
                    ? 'bg-slate-800 border-cyan-500/60 text-cyan-300'
                    : 'bg-slate-900 border-slate-800 text-slate-400 hover:text-slate-200 hover:border-slate-700'
                }`}
              >
                {chip.label}
              </button>
            );
          })}
        </div>

        {/* Live Search Input */}
        <div className="relative min-w-[260px]">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-500">
            <Search className="w-3.5 h-3.5" />
          </div>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => onSearchQueryChange(e.target.value)}
            placeholder="Search by name, phone, or UUID..."
            className="w-full bg-slate-900 border border-slate-800 rounded-lg pl-9 pr-3 py-1.5 text-xs text-white placeholder:text-slate-500 focus:outline-none focus:border-cyan-500 font-mono transition"
          />
        </div>
      </div>
    </div>
  );
};
