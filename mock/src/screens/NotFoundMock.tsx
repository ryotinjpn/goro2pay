import { Link } from 'react-router-dom';
import ScreenFrame from '../components/ScreenFrame';
import BrandHeader from '../components/BrandHeader';
import { COPY } from '../lib/copy';
import styles from './NotFoundMock.module.css';

export default function NotFoundMock() {
  return (
    <ScreenFrame>
      <BrandHeader />
      <div className={styles.center}>
        <h1 className={styles.h1}>{COPY.notFound.h1}</h1>
        <p className={styles.body}>{COPY.notFound.body}</p>
        <Link to="/" className={styles.link}>
          {COPY.notFound.link}
        </Link>
      </div>
    </ScreenFrame>
  );
}
