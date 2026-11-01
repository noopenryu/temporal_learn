import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { Stepper } from '../components/Stepper';

describe('Stepper Component', () => {
  it('renders all steps and displays progress percentage', () => {
    render(
      <Stepper
        currentStep={2}
        completedSteps={['personal_info']}
        progressPercent={20}
      />
    );

    expect(screen.getByText('20%')).toBeInTheDocument();
    expect(screen.getByText('Personal')).toBeInTheDocument();
    expect(screen.getByText('Address')).toBeInTheDocument();
    expect(screen.getByText('Identity')).toBeInTheDocument();
    expect(screen.getByText('Financial')).toBeInTheDocument();
    expect(screen.getByText('Wallet Tier')).toBeInTheDocument();
    expect(screen.getByText('Review')).toBeInTheDocument();
  });
});
