import React from 'react';
import {
  User,
  MapPin,
  FileText,
  BadgeDollarSign,
  CreditCard,
  ClipboardCheck,
  Check,
} from 'lucide-react';

export interface StepItem {
  id: number;
  key: string;
  label: string;
  icon: React.ElementType;
}

export const STEPS: StepItem[] = [
  { id: 1, key: 'personal_info', label: 'Personal', icon: User },
  { id: 2, key: 'address_info', label: 'Address', icon: MapPin },
  { id: 3, key: 'identity_documents', label: 'Identity', icon: FileText },
  { id: 4, key: 'financial_info', label: 'Financial', icon: BadgeDollarSign },
  { id: 5, key: 'wallet_tier', label: 'Wallet Tier', icon: CreditCard },
  { id: 6, key: 'review', label: 'Review', icon: ClipboardCheck },
];

interface StepperProps {
  currentStep: number; // 1 to 6
  completedSteps: string[];
  progressPercent: number;
  onStepClick?: (stepNumber: number) => void;
}

export const Stepper: React.FC<StepperProps> = ({
  currentStep,
  completedSteps,
  progressPercent,
  onStepClick,
}) => {
  return (
    <div className="w-full bg-white border border-slate-200/90 rounded-2xl p-4 sm:p-5 mb-6 shadow-xs">
      {/* Top row: Progress bar and step metadata */}
      <div className="mb-4">
        <div className="flex justify-between items-center text-xs font-semibold mb-2">
          <div className="flex items-center space-x-2">
            <span className="text-slate-700 font-bold uppercase tracking-wider text-[11px]">
              Step {currentStep} of {STEPS.length}
            </span>
          </div>
          <div className="flex items-center space-x-1.5">
            <span className="text-slate-500 text-[11px] font-medium">Completion:</span>
            <span className="text-blue-700 font-semibold font-mono text-sm tabular-nums">
              {progressPercent}%
            </span>
          </div>
        </div>

        {/* Progress rail */}
        <div className="w-full bg-slate-100 rounded-full h-2 overflow-hidden border border-slate-200/40">
          <div
            className="bg-gradient-to-r from-blue-600 to-indigo-600 h-2 rounded-full transition-all duration-500 ease-out origin-left"
            style={{ width: `${Math.max(5, progressPercent)}%` }}
          />
        </div>
      </div>

      {/* Steps indicators */}
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-2 sm:gap-2.5">
        {STEPS.map((step) => {
          const isCompleted = completedSteps.includes(step.key);
          const isCurrent = currentStep === step.id;
          const isClickable = !!onStepClick && (isCompleted || step.id <= currentStep);

          return (
            <button
              key={step.id}
              type="button"
              disabled={!isClickable}
              onClick={() => onStepClick && onStepClick(step.id)}
              className={`flex items-center space-x-2.5 p-2 sm:p-2.5 rounded-xl border text-left min-h-[44px] transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-600 ${
                isCurrent
                  ? 'bg-blue-50/80 border-blue-500/60 text-blue-900 shadow-xs ring-1 ring-blue-500/30'
                  : isCompleted
                  ? 'bg-emerald-50/40 border-emerald-200 text-slate-800 hover:border-emerald-400'
                  : 'bg-slate-50/60 border-slate-200/70 text-slate-400 cursor-not-allowed'
              }`}
            >
              <div
                className={`w-6 h-6 sm:w-7 sm:h-7 rounded-lg flex items-center justify-center shrink-0 text-xs font-bold tabular-nums transition-colors ${
                  isCurrent
                    ? 'bg-blue-600 text-white shadow-2xs shadow-blue-500/30'
                    : isCompleted
                    ? 'bg-emerald-600 text-white'
                    : 'bg-slate-200/80 text-slate-500'
                }`}
              >
                {isCompleted ? <Check className="w-3.5 h-3.5 stroke-[3]" aria-hidden="true" /> : step.id}
              </div>
              <div className="min-w-0 flex-1">
                <p className={`text-xs font-semibold truncate ${isCurrent ? 'text-blue-900' : isCompleted ? 'text-slate-900' : 'text-slate-500'}`}>
                  {step.label}
                </p>
                <p className="text-[10px] text-slate-500 capitalize leading-tight">
                  {isCurrent ? 'Active' : isCompleted ? 'Done' : 'Next'}
                </p>
              </div>
            </button>
          );
        })}
      </div>
    </div>
  );
};
