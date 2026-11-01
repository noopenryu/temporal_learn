import React, { useId, useState } from 'react';
import { BadgeDollarSign, ArrowRight, ArrowLeft, ShieldAlert } from 'lucide-react';
import { FinancialInfo } from '../../types/onboarding';

interface FinancialStepProps {
  initialData?: FinancialInfo;
  onSave: (data: FinancialInfo) => Promise<void>;
  onBack: () => void;
  isLoading: boolean;
}

const SOURCES = [
  'Salary / Employment',
  'Business Profits',
  'Investments / Dividends',
  'Personal Savings',
  'Inheritance / Family Grant',
];

export const FinancialStep: React.FC<FinancialStepProps> = ({
  initialData,
  onSave,
  onBack,
  isLoading,
}) => {
  const occupationId = useId();
  const employerId = useId();
  const incomeId = useId();
  const sourceId = useId();
  const taxIdInput = useId();

  const [occupation, setOccupation] = useState(initialData?.occupation || 'Software Engineer');
  const [employer, setEmployer] = useState(initialData?.employer_name || 'Tech Corp Global');
  const [income, setIncome] = useState(initialData?.annual_income ?? 75000);
  const [sourceOfFunds, setSourceOfFunds] = useState(
    initialData?.source_of_funds || 'Salary / Employment'
  );
  const [declaredPep, setDeclaredPep] = useState(initialData?.declared_pep || false);
  const [taxId, setTaxId] = useState(initialData?.tax_id || 'TAX-99887766');

  const [errors, setErrors] = useState<Record<string, string>>({});

  const validate = (): boolean => {
    const errs: Record<string, string> = {};
    if (!occupation.trim()) {
      errs.occupation = 'Primary occupation is required';
    }
    if (isNaN(Number(income)) || Number(income) < 0) {
      errs.income = 'Annual income must be a valid non-negative number';
    }

    setErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    onSave({
      occupation: occupation.trim(),
      employer_name: employer.trim() || undefined,
      annual_income: Number(income),
      source_of_funds: sourceOfFunds,
      declared_pep: declaredPep,
      tax_id: taxId.trim() || undefined,
    });
  };

  return (
    <div className="bg-white border border-slate-200/90 rounded-2xl p-6 sm:p-8 shadow-xs max-w-2xl mx-auto">
      {/* Header */}
      <div className="flex items-center space-x-3 pb-6 border-b border-slate-100">
        <div className="w-10 h-10 rounded-xl bg-blue-50 border border-blue-200 flex items-center justify-center text-blue-600">
          <BadgeDollarSign className="w-5 h-5" aria-hidden="true" />
        </div>
        <div>
          <h2 className="text-lg font-bold text-slate-900 tracking-tight">Financial & Regulatory Declaration</h2>
          <p className="text-xs text-slate-500">
            Mandated by anti-money laundering (AML) and counter-terrorist financing (CFT) laws.
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="pt-6 space-y-5" noValidate>
        {/* Occupation & Employer */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label htmlFor={occupationId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Primary Occupation <span className="text-rose-500">*</span>
            </label>
            <input
              id={occupationId}
              name="occupation"
              type="text"
              required
              value={occupation}
              onChange={(e) => {
                setOccupation(e.target.value);
                if (errors.occupation) setErrors((prev) => ({ ...prev, occupation: '' }));
              }}
              placeholder="e.g. Software Engineer"
              className={`w-full bg-slate-50 border rounded-xl px-4 py-2.5 text-sm text-slate-900 transition-colors focus:bg-white ${
                errors.occupation ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
              }`}
            />
            {errors.occupation && <p className="text-xs text-rose-600 mt-1 font-medium">{errors.occupation}</p>}
          </div>

          <div>
            <label htmlFor={employerId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Employer / Organization Name
            </label>
            <input
              id={employerId}
              name="organization"
              type="text"
              value={employer}
              onChange={(e) => setEmployer(e.target.value)}
              placeholder="e.g. Tech Corp Global"
              className="w-full bg-slate-50 border border-slate-300 rounded-xl px-4 py-2.5 text-sm text-slate-900 transition-colors focus:bg-white focus:border-blue-600"
            />
          </div>
        </div>

        {/* Estimated Annual Income with Interactive Quick Tiers */}
        <div>
          <div className="flex items-center justify-between mb-1.5">
            <label htmlFor={incomeId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider">
              Estimated Annual Income (USD) <span className="text-rose-500">*</span>
            </label>
            <span className="text-xs font-bold font-mono text-blue-700">
              ${Number(income).toLocaleString('en-US')} / year
            </span>
          </div>
          <div className="space-y-2">
            <input
              id={incomeId}
              type="number"
              required
              min={0}
              step={1000}
              value={income}
              onChange={(e) => {
                setIncome(Number(e.target.value));
                if (errors.income) setErrors((prev) => ({ ...prev, income: '' }));
              }}
              className={`w-full bg-slate-50 border rounded-xl px-4 py-2.5 text-sm font-mono text-slate-900 transition-colors focus:bg-white ${
                errors.income ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
              }`}
            />
            {/* Quick selectors */}
            <div className="flex flex-wrap gap-2 pt-1">
              {[25000, 50000, 75000, 120000, 200000].map((val) => (
                <button
                  key={val}
                  type="button"
                  onClick={() => setIncome(val)}
                  className={`text-xs px-2.5 py-1 rounded-lg border font-mono transition-colors ${
                    income === val
                      ? 'bg-blue-50 border-blue-600 text-blue-800 font-bold'
                      : 'border-slate-200 bg-slate-50 text-slate-600 hover:border-slate-300'
                  }`}
                >
                  ${(val / 1000).toFixed(0)}k
                </button>
              ))}
            </div>
          </div>
          {errors.income && <p className="text-xs text-rose-600 mt-1 font-medium">{errors.income}</p>}
        </div>

        {/* Source of Funds & Tax ID */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label htmlFor={sourceId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Primary Source of Funds <span className="text-rose-500">*</span>
            </label>
            <select
              id={sourceId}
              value={sourceOfFunds}
              onChange={(e) => setSourceOfFunds(e.target.value)}
              className="w-full bg-slate-50 border border-slate-300 rounded-xl px-4 py-2.5 text-sm text-slate-900 transition-colors focus:bg-white focus:border-blue-600"
            >
              {SOURCES.map((src) => (
                <option key={src} value={src}>
                  {src}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label htmlFor={taxIdInput} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Tax ID / TIN / SSN (Optional)
            </label>
            <input
              id={taxIdInput}
              type="text"
              value={taxId}
              onChange={(e) => setTaxId(e.target.value)}
              placeholder="e.g. TAX-99887766"
              className="w-full bg-slate-50 border border-slate-300 rounded-xl px-4 py-2.5 text-sm font-mono text-slate-900 transition-colors focus:bg-white focus:border-blue-600"
            />
          </div>
        </div>

        {/* PEP (Politically Exposed Person) Card with Educational Guidance */}
        <div className={`p-4 rounded-xl border transition-all ${
          declaredPep ? 'bg-amber-50/70 border-amber-300 ring-1 ring-amber-300' : 'bg-slate-50 border-slate-200'
        }`}>
          <div className="flex items-start space-x-3">
            <input
              id="pep-checkbox"
              type="checkbox"
              checked={declaredPep}
              onChange={(e) => setDeclaredPep(e.target.checked)}
              className="mt-1 h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-600"
            />
            <div className="flex-1">
              <label htmlFor="pep-checkbox" className="text-xs font-bold text-slate-900 cursor-pointer block">
                I am a Politically Exposed Person (PEP), or an immediate family member of one
              </label>
              <p className="text-[11px] text-slate-500 mt-0.5 leading-relaxed">
                A PEP is an individual who is or has been entrusted with a prominent public function (e.g. Head of State, senior politician, senior military official, or judiciary).
              </p>
              {declaredPep && (
                <div className="mt-2.5 p-2 rounded-lg bg-amber-100/70 text-amber-900 text-xs font-medium flex items-center space-x-2">
                  <ShieldAlert className="w-4 h-4 text-amber-700 shrink-0" />
                  <span>Your application will be routed to senior compliance for enhanced due diligence (EDD).</span>
                </div>
              )}
            </div>
          </div>
        </div>

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
            <span>{isLoading ? 'Saving...' : 'Continue to Wallet Tier'}</span>
            <ArrowRight className="w-4 h-4" aria-hidden="true" />
          </button>
        </div>
      </form>
    </div>
  );
};
