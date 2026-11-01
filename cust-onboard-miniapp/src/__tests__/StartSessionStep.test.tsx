import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { StartSessionStep } from '../components/steps/StartSessionStep';

describe('StartSessionStep', () => {
  it('renders phone, email, and password fields for new application', () => {
    render(
      <StartSessionStep
        onStart={vi.fn()}
        onResume={vi.fn()}
        isLoading={false}
      />
    );

    // Customer Identifier must NOT be present
    expect(screen.queryByText(/customer identifier/i)).toBeNull();

    // Phone number, Email and Password should be present
    expect(screen.getByLabelText(/mobile phone number/i)).toBeDefined();
    expect(screen.getByLabelText(/email address/i)).toBeDefined();
    expect(screen.getByLabelText(/create application password/i)).toBeDefined();
  });

  it('submits onStart with phone, email, and password', async () => {
    const onStart = vi.fn();
    render(
      <StartSessionStep
        onStart={onStart}
        onResume={vi.fn()}
        isLoading={false}
      />
    );

    const phoneInput = screen.getByLabelText(/mobile phone number/i);
    const emailInput = screen.getByLabelText(/email address/i);
    const passwordInput = screen.getByLabelText(/create application password/i);

    fireEvent.change(phoneInput, { target: { value: '+15559876543' } });
    fireEvent.change(emailInput, { target: { value: 'user@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'Secret123!' } });

    const submitBtn = screen.getByRole('button', { name: /start onboarding/i });
    fireEvent.click(submitBtn);

    expect(onStart).toHaveBeenCalledWith({
      phone_number: '+15559876543',
      email: 'user@example.com',
      password: 'Secret123!',
    });
  });

  it('resumes by phone number AND password in the resume tab', async () => {
    const onResume = vi.fn();
    render(
      <StartSessionStep
        onStart={vi.fn()}
        onResume={onResume}
        isLoading={false}
      />
    );

    // Switch to Resume tab
    const resumeTabBtn = screen.getByRole('button', { name: /^resume application$/i });
    fireEvent.click(resumeTabBtn);

    const phoneInput = screen.getByLabelText(/registered phone number/i);
    const passwordInput = screen.getByLabelText(/application password/i);

    fireEvent.change(phoneInput, { target: { value: '+15559876543' } });
    fireEvent.change(passwordInput, { target: { value: 'MySecretPass!' } });

    const resumeSubmitBtn = screen.getByRole('button', { name: /resume my application/i });
    fireEvent.click(resumeSubmitBtn);

    expect(onResume).toHaveBeenCalledWith({
      phone_number: '+15559876543',
      password: 'MySecretPass!',
    });
  });

  it('displays error notice when error prop is provided', () => {
    render(
      <StartSessionStep
        onStart={vi.fn()}
        onResume={vi.fn()}
        isLoading={false}
        error="Invalid phone number or password. Please verify and try again."
      />
    );

    expect(screen.getByText(/unable to proceed/i)).toBeDefined();
    expect(screen.getByText(/invalid phone number or password/i)).toBeDefined();
  });
});
