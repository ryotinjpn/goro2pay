import { FormEvent, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ScreenFrame from '../components/ScreenFrame';
import BrandHeader from '../components/BrandHeader';
import { COPY } from '../lib/copy';
import styles from './LoginMock.module.css';

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export default function LoginMock() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const isEmailValid = EMAIL_REGEX.test(email);
  const canSubmit = isEmailValid && password.length > 0 && !submitting;

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
      <h1 className={styles.h1}>{COPY.login.h1}</h1>
      <div className={styles.sub}>{COPY.login.sub}</div>
      <form onSubmit={handleSubmit} className={styles.form}>
        <div className={styles.field}>
          <label htmlFor="login-email">{COPY.login.emailLabel}</label>
          <input
            id="login-email"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        <div className={styles.field}>
          <label htmlFor="login-password">{COPY.login.passwordLabel}</label>
          <input
            id="login-password"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>
        <button
          type="submit"
          className={styles.submit}
          disabled={!canSubmit}
        >
          {submitting ? COPY.login.submitting : COPY.login.submit}
        </button>
      </form>
    </ScreenFrame>
  );
}
