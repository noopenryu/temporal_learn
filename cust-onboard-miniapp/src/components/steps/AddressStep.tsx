import React, { useId, useState } from 'react';
import { MapPin, ArrowRight, ArrowLeft, FileText, CheckCircle2, UploadCloud } from 'lucide-react';
import { AddressInfo } from '../../types/onboarding';

interface AddressStepProps {
  initialData?: AddressInfo;
  onSave: (data: AddressInfo) => Promise<void>;
  onBack: () => void;
  isLoading: boolean;
}

const COUNTRIES = [
  { code: 'VN', name: 'Vietnam', flag: '🇻🇳' },
  { code: 'SG', name: 'Singapore', flag: '🇸🇬' },
  { code: 'US', name: 'United States', flag: '🇺🇸' },
  { code: 'GB', name: 'United Kingdom', flag: '🇬🇧' },
  { code: 'ID', name: 'Indonesia', flag: '🇮🇩' },
  { code: 'MY', name: 'Malaysia', flag: '🇲🇾' },
  { code: 'TH', name: 'Thailand', flag: '🇹🇭' },
  { code: 'JP', name: 'Japan', flag: '🇯🇵' },
  { code: 'DE', name: 'Germany', flag: '🇩🇪' },
  { code: 'FR', name: 'France', flag: '🇫🇷' },
  { code: 'AU', name: 'Australia', flag: '🇦🇺' },
];

export const AddressStep: React.FC<AddressStepProps> = ({
  initialData,
  onSave,
  onBack,
  isLoading,
}) => {
  const countryId = useId();
  const stateId = useId();
  const cityId = useId();
  const postalCodeId = useId();
  const streetAddressId = useId();

  const [country, setCountry] = useState(initialData?.country || 'VN');
  const [state, setState] = useState(initialData?.state || 'Hanoi');
  const [city, setCity] = useState(initialData?.city || 'Hanoi');
  const [streetAddress, setStreetAddress] = useState(
    initialData?.street_address || '123 Tran Hung Dao Street'
  );
  const [postalCode, setPostalCode] = useState(initialData?.postal_code || '100000');
  const [proofUrl, setProofUrl] = useState(
    initialData?.proof_of_address_url || 'https://documents.wallet.internal/proof/utility-bill.pdf'
  );
  const [uploadMockState, setUploadMockState] = useState<'sample' | 'custom'>('sample');

  const [errors, setErrors] = useState<Record<string, string>>({});

  const validate = (): boolean => {
    const errs: Record<string, string> = {};
    if (!country.trim() || country.trim().length < 2) {
      errs.country = 'Country code is required (e.g. VN, US, SG)';
    }
    if (!city.trim() || city.trim().length < 2) {
      errs.city = 'City must be at least 2 characters';
    }
    if (!postalCode.trim() || postalCode.trim().length < 3) {
      errs.postalCode = 'Postal code must be at least 3 characters';
    }
    if (!streetAddress.trim() || streetAddress.trim().length < 5) {
      errs.streetAddress = 'Street address must be at least 5 characters';
    }

    setErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    onSave({
      country: country.trim(),
      state: state.trim(),
      city: city.trim(),
      street_address: streetAddress.trim(),
      postal_code: postalCode.trim(),
      proof_of_address_url: proofUrl.trim(),
    });
  };

  return (
    <div className="bg-white border border-slate-200/90 rounded-2xl p-6 sm:p-8 shadow-xs max-w-2xl mx-auto">
      {/* Header */}
      <div className="flex items-center space-x-3 pb-6 border-b border-slate-100">
        <div className="w-10 h-10 rounded-xl bg-blue-50 border border-blue-200 flex items-center justify-center text-blue-600">
          <MapPin className="w-5 h-5" aria-hidden="true" />
        </div>
        <div>
          <h2 className="text-lg font-bold text-slate-900 tracking-tight">Residential Address</h2>
          <p className="text-xs text-slate-500">Required for regulatory residency confirmation and tax jurisdiction.</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="pt-6 space-y-5" noValidate>
        {/* Country & Province/State */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label htmlFor={countryId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Country <span className="text-rose-500">*</span>
            </label>
            <select
              id={countryId}
              name="country"
              autoComplete="country"
              value={country}
              onChange={(e) => {
                setCountry(e.target.value);
                if (errors.country) setErrors((prev) => ({ ...prev, country: '' }));
              }}
              className="w-full bg-slate-50 border border-slate-300 rounded-xl px-4 py-2.5 text-sm text-slate-900 transition-colors focus:bg-white focus:border-blue-600"
            >
              {COUNTRIES.map((c) => (
                <option key={c.code} value={c.code}>
                  {c.flag} {c.name} ({c.code})
                </option>
              ))}
            </select>
            {errors.country && <p className="text-xs text-rose-600 mt-1 font-medium">{errors.country}</p>}
          </div>

          <div>
            <label htmlFor={stateId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              State / Province / Region
            </label>
            <input
              id={stateId}
              name="address-level1"
              autoComplete="address-level1"
              type="text"
              value={state}
              onChange={(e) => setState(e.target.value)}
              placeholder="e.g. Hanoi, California"
              className="w-full bg-slate-50 border border-slate-300 rounded-xl px-4 py-2.5 text-sm text-slate-900 transition-colors focus:bg-white focus:border-blue-600"
            />
          </div>
        </div>

        {/* City & Postal Code */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label htmlFor={cityId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              City <span className="text-rose-500">*</span>
            </label>
            <input
              id={cityId}
              name="address-level2"
              autoComplete="address-level2"
              type="text"
              required
              value={city}
              onChange={(e) => {
                setCity(e.target.value);
                if (errors.city) setErrors((prev) => ({ ...prev, city: '' }));
              }}
              placeholder="e.g. Hanoi"
              className={`w-full bg-slate-50 border rounded-xl px-4 py-2.5 text-sm text-slate-900 transition-colors focus:bg-white ${
                errors.city ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
              }`}
            />
            {errors.city && <p className="text-xs text-rose-600 mt-1 font-medium">{errors.city}</p>}
          </div>

          <div>
            <label htmlFor={postalCodeId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Postal / ZIP Code <span className="text-rose-500">*</span>
            </label>
            <input
              id={postalCodeId}
              name="postal-code"
              autoComplete="postal-code"
              type="text"
              required
              value={postalCode}
              onChange={(e) => {
                setPostalCode(e.target.value);
                if (errors.postalCode) setErrors((prev) => ({ ...prev, postalCode: '' }));
              }}
              placeholder="e.g. 100000"
              className={`w-full bg-slate-50 border rounded-xl px-4 py-2.5 text-sm text-slate-900 font-mono transition-colors focus:bg-white ${
                errors.postalCode ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
              }`}
            />
            {errors.postalCode && <p className="text-xs text-rose-600 mt-1 font-medium">{errors.postalCode}</p>}
          </div>
        </div>

        {/* Street Address */}
        <div>
          <label htmlFor={streetAddressId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
            Street Address (Building, Suite, Street) <span className="text-rose-500">*</span>
          </label>
          <input
            id={streetAddressId}
            name="street-address"
            autoComplete="street-address"
            type="text"
            required
            value={streetAddress}
            onChange={(e) => {
              setStreetAddress(e.target.value);
              if (errors.streetAddress) setErrors((prev) => ({ ...prev, streetAddress: '' }));
            }}
            placeholder="e.g. 123 Tran Hung Dao Street, Floor 4"
            className={`w-full bg-slate-50 border rounded-xl px-4 py-2.5 text-sm text-slate-900 transition-colors focus:bg-white ${
              errors.streetAddress ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
            }`}
          />
          {errors.streetAddress && <p className="text-xs text-rose-600 mt-1 font-medium">{errors.streetAddress}</p>}
        </div>

        {/* Proof of Address Upload Section */}
        <div className="pt-2">
          <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
            Proof of Address Document (Optional)
          </label>
          <div className="p-4 rounded-xl border border-dashed border-slate-300 bg-slate-50/70 hover:bg-slate-50 transition-colors">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <div className="w-10 h-10 rounded-lg bg-blue-100/60 text-blue-700 flex items-center justify-center">
                  <FileText className="w-5 h-5" />
                </div>
                <div>
                  <p className="text-xs font-semibold text-slate-800">
                    {uploadMockState === 'sample' ? 'Sample Utility Statement Verified' : 'Custom Document Uploaded'}
                  </p>
                  <p className="text-[11px] text-slate-500">Bank Statement, Utility Bill, or Rental Contract (&lt; 3 months)</p>
                </div>
              </div>
              <span className="inline-flex items-center space-x-1 px-2.5 py-1 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200 text-[11px] font-semibold">
                <CheckCircle2 className="w-3.5 h-3.5 text-emerald-600" />
                <span>Verified</span>
              </span>
            </div>

            <div className="mt-3 pt-3 border-t border-slate-200/60 flex items-center justify-between text-xs">
              <span className="font-mono text-[11px] text-slate-500 truncate max-w-xs">{proofUrl}</span>
              <button
                type="button"
                onClick={() => {
                  setProofUrl('https://documents.wallet.internal/proof/residence-cert-2026.pdf');
                  setUploadMockState('custom');
                }}
                className="text-blue-600 hover:text-blue-700 font-semibold text-xs flex items-center space-x-1"
              >
                <UploadCloud className="w-3.5 h-3.5" />
                <span>Replace file</span>
              </button>
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
            <span>{isLoading ? 'Saving...' : 'Continue to Identity'}</span>
            <ArrowRight className="w-4 h-4" aria-hidden="true" />
          </button>
        </div>
      </form>
    </div>
  );
};
