import styles from './SuggestBubble.module.css';

type Props = {
  text: string;
};

export default function SuggestBubble({ text }: Props) {
  return <div className={styles.bubble}>{text}</div>;
}
