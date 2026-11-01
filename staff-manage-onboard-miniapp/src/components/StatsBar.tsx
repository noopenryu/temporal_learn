import React from 'react';
import { ApplicationDetailResponse } from '../types/staff';

interface StatsBarProps {
  items: ApplicationDetailResponse[];
  onSelectFilter: (status: string) => void;
  activeStatus: string;
}

export const StatsBar: React.FC<StatsBarProps> = ({
  items,
  onSelectFilter,
  activeStatus,
}) => {
  const manualReviewCount = items.filter(
    (i) => i.application.status === 'REQUIRES_MANUAL_REVIEW'
  ).length;
  const processingCount = items.filter(
    (i) => i.application.status === 'PROCESSING' || i.application.status === 'SUBMITTED'
  ).length;
  const approvedCount = items.filter(
    (i) => i.application.status === 'APPROVED'
  ).length;
  const rejectedCount = items.filter(
    (i) => i.application.status === 'REJECTED'
  ).length;
  const confirmedCount = items.filter(
    (i) => !!i.application.confirmed_data
  ).length;

  return (
    <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
      {/* 1. Pending Manual Review */}
      <button
        type="button"
        onClick={() => onSelectFilter(activeStatus === 'REQUIRES_MANUAL_REVIEW' ? 'ALL' : 'REQUIRES_MANUAL_REVIEW')}
        className={`text-left p-3.5 rounded-xl border transition-all ${
          activeStatus === 'REQUIRES_MANUAL_REVIEW'
            ? 'bg-amber-950/40 border-amber-500/70 ring-1 ring-amber-500/30'
            : 'bg-slate-900 border-slate-800 hover:border-slate-700'
        }`}
      >
        <div className="flex items-center justify-between">
          <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Needs Review</span>
          {manualReviewCount > 0 && (
            <span className="w-2 h-2 rounded-full bg-amber-400" />
          )}
        </div>
        <div className="mt-2 flex items-baseline space-x-2">
          <span className="text-2xl font-bold text-white font-mono">{manualReviewCount}</span>
          {manualReviewCount > 0 && (
            <span className="text-[10px] text-amber-400 font-medium">Action Required</span>
          )}
        </div>
      </button>

      {/* 2. In-Flight Workflows */}
      <button
        type="button"
        onClick={() => onSelectFilter(activeStatus === 'PROCESSING' ? 'ALL' : 'PROCESSING')}
        className={`text-left p-3.5 rounded-xl border transition-all ${
          activeStatus === 'PROCESSING'
            ? 'bg-blue-950/40 border-blue-500/70 ring-1 ring-blue-500/30'
            : 'bg-slate-900 border-slate-800 hover:border-slate-700'
        }`}
      >
        <div className="flex items-center justify-between">
          <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">In-Flight / Sagas</span>
          {processingCount > 0 && (
            <span className="w-2 h-2 rounded-full bg-blue-400" />
          )}
        </div>
        <div className="mt-2 flex items-baseline space-x-2">
          <span className="text-2xl font-bold text-white font-mono">{processingCount}</span>
          <span className="text-[10px] text-slate-500 font-mono">Temporal</span>
        </div>
      </button>

      {/* 3. Approved */}
      <button
        type="button"
        onClick={() => onSelectFilter(activeStatus === 'APPROVED' ? 'ALL' : 'APPROVED')}
        className={`text-left p-3.5 rounded-xl border transition-all ${
          activeStatus === 'APPROVED'
            ? 'bg-emerald-950/40 border-emerald-500/70 ring-1 ring-emerald-500/30'
            : 'bg-slate-900 border-slate-800 hover:border-slate-700'
        }`}
      >
        <div className="flex items-center justify-between">
          <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Approved</span>
          <span className="text-[10px] font-mono text-emerald-400 font-medium">Passed</span>
        </div>
        <div className="mt-2">
          <span className="text-2xl font-bold text-white font-mono">{approvedCount}</span>
        </div>
      </button>

      {/* 4. Rejected */}
      <button
        type="button"
        onClick={() => onSelectFilter(activeStatus === 'REJECTED' ? 'ALL' : 'REJECTED')}
        className={`text-left p-3.5 rounded-xl border transition-all ${
          activeStatus === 'REJECTED'
            ? 'bg-rose-950/40 border-rose-500/70 ring-1 ring-rose-500/30'
            : 'bg-slate-900 border-slate-800 hover:border-slate-700'
        }`}
      >
        <div className="flex items-center justify-between">
          <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Rejected</span>
          <span className="text-[10px] font-mono text-rose-400 font-medium">Blocked</span>
        </div>
        <div className="mt-2">
          <span className="text-2xl font-bold text-white font-mono">{rejectedCount}</span>
        </div>
      </button>

      {/* 5. Confirmed Data Overlays */}
      <div className="p-3.5 rounded-xl border border-slate-800 bg-slate-900/60 flex flex-col justify-between">
        <div className="flex items-center justify-between">
          <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Staff Confirmed</span>
          <span className="text-[10px] font-mono text-cyan-400 font-medium">Audited</span>
        </div>
        <div className="mt-2">
          <span className="text-2xl font-bold text-white font-mono">{confirmedCount}</span>
        </div>
      </div>
    </div>
  );
};
