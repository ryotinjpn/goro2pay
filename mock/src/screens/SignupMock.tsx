import { FormEvent, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ScreenFrame from '../components/ScreenFrame';
import BrandHeader from '../components/BrandHeader';
import { COPY } from '../lib/copy';
import styles from './SignupMock.module.css';

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

type PwdHints = {
  lengthOk: boolean;
  upperOk: boolean;
  lowerOk: boolean;
  digitOk: boolean;
};

function checkPasswordHints(pw: string): PwdHints {
  return {
    lengthOk: pw.length >= 8,
    upperOk: /[A-Z]/.test(pw),
    lowerOk: /[a-z]/.test(pw),
    digitOk: /\d/.test(pw),
  };
}

function isPasswordValid(h: PwdHints): boolean {
  return h.lengthOk && h.upperOk && h.lowerOk && h.digitOk;
}

export default function SignupMock() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const hints = checkPasswordHints(password);
  const isEmailValid = EMAIL_REGEX.test(email);
  const canSubmit = isEmailValid && isPasswordValid(hints) && !submitting;

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    setSubmitting(true);
    setTimeout(() => {
      setSubmitting(false);
      navigate('/main');
    }, 800);
  };

  return (
    <ScreenFrame>
      <BrandHeader />
      <h1 className={styles.h1}>
        <div>{COPY.signup.h1Line1}</div>
        <div>{COPY.signup.h1Line2}</div>
      </h1>
      <div className={styles.sub}>{COPY.signup.sub}</div>
      <form onSubmit={handleSubmit} className={styles.form}>
        <div className={styles.field}>
          <label htmlFor="signup-email">{COPY.signup.emailLabel}</label>
          <input
            id="signup-email"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        <div className={styles.field}>
          <label htmlFor="signup-password">{COPY.signup.passwordLabel}</label>
          <input
            id="signup-password"
            type="password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          <ul className={styles.hints} aria-live="polite">
            <li className={hints.lengthOk ? styles.ok : ''}>
              {hints.lengthOk ? '✓' : '・'} {COPY.signup.hintLength}
            </li>
            <li className={hints.upperOk ? styles.ok : ''}>
              {hints.upperOk ? '✓' : '・'} {COPY.signup.hintUpper}
            </li>
            <li className={hints.lowerOk ? styles.ok : ''}>
              {hints.lowerOk ? '✓' : '・'} {COPY.signup.hintLower}
            </li>
            <li className={hints.digitOk ? styles.ok : ''}>
              {hints.digitOk ? '✓' : '・'} {COPY.signup.hintDigit}
            </li>
          </ul>
        </div>
        <button
          type="submit"
          className={styles.submit}
          disabled={!canSubmit}
        >
          {submitting ? COPY.signup.submitting : COPY.signup.submit}
        </button>
      </form>
    </ScreenFrame>
  );
}
