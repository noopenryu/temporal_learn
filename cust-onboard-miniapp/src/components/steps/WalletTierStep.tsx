import React, { useState } from 'react';
import { CreditCard, Check, ArrowRight, ArrowLeft, Sparkles, Globe } from 'lucide-react';
import { WalletTier } from '../../types/onboarding';

interface WalletTierStepProps {
  initialTier?: string;
  onSave: (tier: WalletTier) => Promise<void>;
  onBack: () => void;
  isLoading: boolean;
}

export const WalletTierStep: React.FC<WalletTierStepProps> = ({
  initialTier,
  onSave,
  onBack,
  isLoading,
}) => {
  const [selectedTier, setSelectedTier] = useState<WalletTier>(
    (initialTier as WalletTier) || 'TIER_1_BASIC'
  );

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(selectedTier);
  };

  return (
    <div className="bg-white border border-slate-200/90 rounded-2xl p-6 sm:p-8 shadow-xs max-w-3xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-center space-x-3 pb-6 border-b border-slate-100">
        <div className="w-10 h-10 rounded-xl bg-blue-50 border border-blue-200 flex items-center justify-center text-blue-600">
          <CreditCard className="w-5 h-5" aria-hidden="true" />
        </div>
        <div>
          <h2 className="text-lg font-bold text-slate-900 tracking-tight">Select Wallet Tier</h2>
          <p className="text-xs text-slate-500">
            Choose the account tier that suits your daily transactions, storage limits, and cross-border needs.
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <fieldset>
          <legend className="sr-only">Choose your digital wallet tier</legend>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-5">
            {/* Tier 1 Basic Card */}
            <button
              type="button"
              role="radio"
              aria-checked={selectedTier === 'TIER_1_BASIC'}
              onClick={() => setSelectedTier('TIER_1_BASIC')}
              className={`relative p-5 sm:p-6 rounded-2xl border-2 text-left flex flex-col justify-between transition-all ${
                selectedTier === 'TIER_1_BASIC'
                  ? 'border-blue-600 bg-blue-50/50 shadow-sm ring-1 ring-blue-600'
                  : 'border-slate-200 bg-white hover:border-slate-300'
              }`}
            >
              <div>
                <div className="flex justify-between items-start mb-3">
                  <span className="text-[11px] font-bold px-2.5 py-0.5 rounded-full bg-slate-100 text-slate-700">
                    BASIC ACCOUNT
                  </span>
                  <div
                    className={`w-6 h-6 rounded-full flex items-center justify-center border transition-colors ${
                      selectedTier === 'TIER_1_BASIC'
                        ? 'bg-blue-600 border-blue-600 text-white'
                        : 'border-slate-300 bg-white'
                    }`}
                  >
                    {selectedTier === 'TIER_1_BASIC' && <Check className="w-3.5 h-3.5 stroke-[3]" aria-hidden="true" />}
                  </div>
                </div>

                <h3 className="text-lg font-bold text-slate-900">Tier 1: Basic Wallet</h3>
                <p className="text-xs text-slate-500 mt-1 leading-relaxed">
                  Everyday digital wallet for peer-to-peer payments, local QR code scanning, and daily spending.
                </p>

                {/* Capacity breakdown */}
                <div className="mt-4 pt-4 border-t border-slate-100 space-y-2.5 text-xs">
                  <div>
                    <div className="flex justify-between text-slate-600 mb-1">
                      <span>Maximum Stored Balance:</span>
                      <span className="font-bold text-slate-900 font-mono">$5,000 USD</span>
                    </div>
                    <div className="w-full bg-slate-100 rounded-full h-1.5 overflow-hidden">
                      <div className="bg-slate-400 h-1.5 rounded-full w-1/4" />
                    </div>
                  </div>

                  <div className="flex justify-between text-slate-600 pt-1">
                    <span>Daily Transaction Limit:</span>
                    <span className="font-bold text-slate-900 font-mono">$1,000 USD</span>
                  </div>

                  <div className="flex justify-between text-slate-600">
                    <span>KYC Clearance:</span>
                    <span className="font-semibold text-emerald-700">Instant Automated ID Check</span>
                  </div>
                </div>
              </div>

              <div className="mt-6 pt-4 border-t border-slate-100 flex items-center space-x-2 text-xs text-slate-600">
                <Check className="w-4 h-4 text-emerald-600 shrink-0" aria-hidden="true" />
                <span>Zero monthly account maintenance fee</span>
              </div>
            </button>

            {/* Tier 2 Premium Card */}
            <button
              type="button"
              role="radio"
              aria-checked={selectedTier === 'TIER_2_PREMIUM'}
              onClick={() => setSelectedTier('TIER_2_PREMIUM')}
              className={`relative p-5 sm:p-6 rounded-2xl border-2 text-left flex flex-col justify-between transition-all ${
                selectedTier === 'TIER_2_PREMIUM'
                  ? 'border-blue-600 bg-blue-50/70 shadow-md ring-1 ring-blue-600'
                  : 'border-slate-200 bg-white hover:border-slate-300'
              }`}
            >
              {/* Popular badge */}
              <div className="absolute -top-3 right-6 bg-gradient-to-r from-blue-600 to-indigo-600 text-white text-[10px] font-bold px-2.5 py-0.5 rounded-full shadow-xs flex items-center space-x-1">
                <Sparkles className="w-3 h-3" />
                <span>RECOMMENDED</span>
              </div>

              <div>
                <div className="flex justify-between items-start mb-3">
                  <span className="text-[11px] font-bold px-2.5 py-0.5 rounded-full bg-blue-100 text-blue-800">
                    PREMIUM UNLIMITED
                  </span>
                  <div
                    className={`w-6 h-6 rounded-full flex items-center justify-center border transition-colors ${
                      selectedTier === 'TIER_2_PREMIUM'
                        ? 'bg-blue-600 border-blue-600 text-white'
                        : 'border-slate-300 bg-white'
                    }`}
                  >
                    {selectedTier === 'TIER_2_PREMIUM' && <Check className="w-3.5 h-3.5 stroke-[3]" aria-hidden="true" />}
                  </div>
                </div>

                <h3 className="text-lg font-bold text-slate-900">Tier 2: Premium Wallet</h3>
                <p className="text-xs text-slate-500 mt-1 leading-relaxed">
                  High-limit wallet with global multi-currency accounts, virtual Visa/Mastercard, and cross-border transfers.
                </p>

                {/* Capacity breakdown */}
                <div className="mt-4 pt-4 border-t border-slate-100 space-y-2.5 text-xs">
                  <div>
                    <div className="flex justify-between text-slate-600 mb-1">
                      <span>Maximum Stored Balance:</span>
                      <span className="font-bold text-blue-900 font-mono">$50,000 USD</span>
                    </div>
                    <div className="w-full bg-slate-100 rounded-full h-1.5 overflow-hidden">
                      <div className="bg-gradient-to-r from-blue-600 to-indigo-600 h-1.5 rounded-full w-full" />
                    </div>
                  </div>

                  <div className="flex justify-between text-slate-600 pt-1">
                    <span>Daily Transaction Limit:</span>
                    <span className="font-bold text-blue-900 font-mono">$10,000 USD</span>
                  </div>

                  <div className="flex justify-between text-slate-600">
                    <span>Global Capabilities:</span>
                    <span className="font-semibold text-blue-700">Multi-Currency & Virtual Card</span>
                  </div>
                </div>
              </div>

              <div className="mt-6 pt-4 border-t border-slate-100 flex items-center space-x-2 text-xs text-blue-900 font-medium">
                <Globe className="w-4 h-4 text-blue-600 shrink-0" aria-hidden="true" />
                <span>Zero international FX exchange markup</span>
              </div>
            </button>
          </div>
        </fieldset>

        {/* Controls */}
        <div className="pt-6 flex items-center justify-between border-t border-slate-100">
          <button
            type="button"
            onClick={onBack}
            className="flex items-center space-x-2 px-4 py-2.5 rounded-xl border border-slate-200 text-slate-700 hover:bg-slate-50 font-semibold text-xs transition-colors"
          >
            <ArrowLeft className="w-4 h-4" aria-hidden="true" />
            <span>Back</span>
          </button>

          <button
            type="submit"
            disabled={isLoading}
            className="flex items-center space-x-2 px-6 py-2.5 rounded-xl bg-blue-600 hover:bg-blue-700 text-white font-semibold text-xs shadow-md shadow-blue-500/20 transition-all disabled:opacity-50"
          >
            <span>{isLoading ? 'Saving...' : 'Save & Review'}</span>
            <ArrowRight className="w-4 h-4" aria-hidden="true" />
          </button>
        </div>
      </form>
    </div>
  );
};
