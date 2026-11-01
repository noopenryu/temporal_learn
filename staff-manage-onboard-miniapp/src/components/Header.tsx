import React from 'react';
import { UserCheck, RefreshCw, Terminal } from 'lucide-react';

interface HeaderProps {
  reviewerId: string;
  onReviewerIdChange: (newId: string) => void;
  onRefresh: () => void;
  isRefreshing: boolean;
}

export const Header: React.FC<HeaderProps> = ({
  reviewerId,
  onReviewerIdChange,
  onRefresh,
  isRefreshing,
}) => {
  return (
    <header className="bg-slate-900 border-b border-slate-800 sticky top-0 z-30 px-4 sm:px-6 py-3">
      <div className="max-w-7xl mx-auto flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        {/* Brand & Portal Title */}
        <div className="flex items-center space-x-3">
          <div className="w-9 h-9 rounded-lg bg-slate-800 border border-slate-700 flex items-center justify-center text-cyan-400 font-mono font-bold text-sm shadow-xs">
            <Terminal className="w-4 h-4" />
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <h1 className="text-sm sm:text-base font-bold text-white tracking-tight">WalletOps Workbench</h1>
              <span className="text-[10px] font-semibold tracking-wider uppercase px-2 py-0.5 rounded bg-amber-500/10 text-amber-300 border border-amber-500/20">
                Staff Review Portal
              </span>
            </div>
            <p className="text-[11px] text-slate-400">
              Customer Onboarding &amp; KYC Manual Case Verification Console
            </p>
          </div>
        </div>

        {/* Right side: Reviewer ID & Actions */}
        <div className="flex items-center space-x-2.5 w-full sm:w-auto justify-between sm:justify-end">
          {/* Reviewer Badge */}
          <div className="flex items-center space-x-2 bg-slate-800/90 border border-slate-700 rounded-lg px-2.5 py-1.5">
            <UserCheck className="w-3.5 h-3.5 text-cyan-400 shrink-0" />
            <div className="text-[11px] flex items-center space-x-1">
              <span className="text-slate-400">Officer:</span>
              <input
                type="text"
                value={reviewerId}
                onChange={(e) => onReviewerIdChange(e.target.value)}
                className="bg-transparent text-white font-mono font-medium focus:outline-none focus:ring-1 focus:ring-cyan-500 rounded px-1 w-36 text-[11px]"
                title="Edit staff reviewer identifier"
              />
            </div>
          </div>

          {/* Temporal & Services Badge */}
          <div className="hidden lg:flex items-center space-x-2 text-[11px] text-slate-400 bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1.5 font-mono">
            <span className="w-2 h-2 rounded-full bg-emerald-400" />
            <span>Temporal Engine: Active</span>
          </div>

          {/* Refresh Button */}
          <button
            type="button"
            onClick={onRefresh}
            disabled={isRefreshing}
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold border border-slate-700 transition disabled:opacity-50"
          >
            <RefreshCw className={`w-3.5 h-3.5 text-cyan-400 ${isRefreshing ? 'animate-spin' : ''}`} />
            <span>{isRefreshing ? 'Syncing...' : 'Sync Queue'}</span>
          </button>
        </div>
      </div>
    </header>
  );
};
