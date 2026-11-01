import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { PersonalInfoStep } from '../components/steps/PersonalInfoStep';

describe('PersonalInfoStep', () => {
  it('does NOT default date of birth to year 0001 when raw DOB has zero-time', () => {
    render(
      <PersonalInfoStep
        initialData={{
          first_name: 'Alice',
          last_name: 'Smith',
          date_of_birth: '0001-01-01T00:00:00Z',
          nationality: 'US',
          phone_number: '+15551234567',
          email: 'alice@example.com',
        }}
        onSave={vi.fn()}
        onBack={vi.fn()}
        isLoading={false}
      />
    );

    const dobInput = screen.getByLabelText(/date of birth/i) as HTMLInputElement;
    expect(dobInput.value).not.toContain('0001');
    expect(Number(dobInput.value.split('-')[0])).toBeGreaterThanOrEqual(1900);
  });

  it('displays underage indicator and disables progression when applicant is under 18', () => {
    const today = new Date();
    const underageDate = new Date(today.getFullYear() - 16, today.getMonth(), today.getDate())
      .toISOString()
      .split('T')[0];

    render(
      <PersonalInfoStep
        initialData={{
          first_name: 'Teen',
          last_name: 'User',
          date_of_birth: underageDate,
          nationality: 'US',
          phone_number: '+15551234567',
          email: 'teen@example.com',
        }}
        onSave={vi.fn()}
        onBack={vi.fn()}
        isLoading={false}
      />
    );

    expect(screen.getByText(/underage <18/i)).toBeDefined();
    const submitBtn = screen.getByRole('button', { name: /continue to address/i });
    expect(submitBtn).toBeDisabled();
  });

  it('submits valid adult details successfully', () => {
    const onSave = vi.fn();
    const adultDob = '1995-06-15';

    render(
      <PersonalInfoStep
        initialData={{
          first_name: 'Alice',
          last_name: 'Smith',
          date_of_birth: adultDob,
          nationality: 'US',
          phone_number: '+15551234567',
          email: 'alice@example.com',
        }}
        onSave={onSave}
        onBack={vi.fn()}
        isLoading={false}
      />
    );

    const submitBtn = screen.getByRole('button', { name: /continue to address/i });
    expect(submitBtn).not.toBeDisabled();
    fireEvent.click(submitBtn);

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        first_name: 'Alice',
        last_name: 'Smith',
        date_of_birth: '1995-06-15',
        phone_number: '+15551234567',
        email: 'alice@example.com',
      })
    );
  });
});
