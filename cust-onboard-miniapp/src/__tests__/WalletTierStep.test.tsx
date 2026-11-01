import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { WalletTierStep } from '../components/steps/WalletTierStep';

describe('WalletTierStep Component', () => {
  it('allows selecting Tier 2 Premium and saves', async () => {
    const handleSave = vi.fn().mockResolvedValue(undefined);
    const handleBack = vi.fn();

    render(
      <WalletTierStep
        initialTier="TIER_1_BASIC"
        onSave={handleSave}
        onBack={handleBack}
        isLoading={false}
      />
    );

    expect(screen.getByText('Tier 1: Basic Wallet')).toBeInTheDocument();
    expect(screen.getByText('Tier 2: Premium Wallet')).toBeInTheDocument();

    // Click Tier 2 Premium
    fireEvent.click(screen.getByText('Tier 2: Premium Wallet'));

    // Click submit
    fireEvent.click(screen.getByText('Save & Review'));

    expect(handleSave).toHaveBeenCalledWith('TIER_2_PREMIUM');
  });
});
