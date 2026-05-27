"use client";

import { useEffect, useState } from "react";

import { COPY } from "@/lib/copy";

import styles from "./BrandHeader.module.css";

function formatTime(d: Date): string {
  const hh = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  return `${hh}:${mm}`;
}

export function BrandHeader() {
  const [time, setTime] = useState<string>(() => formatTime(new Date()));

  useEffect(() => {
    const id = setInterval(() => setTime(formatTime(new Date())), 60_000);
    return () => clearInterval(id);
  }, []);

  const name = COPY.brand.name;
  const splitAt = name.indexOf("Pay");
  const left = splitAt >= 0 ? name.slice(0, splitAt) : name;
  const right = splitAt >= 0 ? name.slice(splitAt) : "";

  return (
    <header className={styles.header}>
      <span className={styles.brand}>
        {left}
        {right && <span className={styles.pay}>{right}</span>}
      </span>
      <span className={styles.time}>{time}</span>
    </header>
  );
}
