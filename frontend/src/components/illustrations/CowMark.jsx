// Logo kepala sapi sederhana (orisinal)
export default function CowMark({ size = 40, className = '' }) {
  return (
    <svg className={'cow-mark ' + className} viewBox="0 0 48 48" width={size} height={size} aria-hidden="true">
      <circle cx="24" cy="24" r="23" fill="var(--color-accent)" />
      <ellipse cx="10" cy="20" rx="6.5" ry="3.4" transform="rotate(-20 10 20)" fill="var(--cow-white)" stroke="var(--cow-outline)" strokeWidth="1.6" />
      <ellipse cx="38" cy="20" rx="6.5" ry="3.4" transform="rotate(20 38 20)" fill="var(--cow-white)" stroke="var(--cow-outline)" strokeWidth="1.6" />
      <path d="M17.5 12.5c-2-3-1.6-5.6.6-7 .2 2.2 1 4 2.6 5.4zM30.5 12.5c2-3 1.6-5.6-.6-7-.2 2.2-1 4-2.6 5.4z" fill="var(--cow-horn)" stroke="var(--cow-outline)" strokeWidth="1.4" strokeLinejoin="round" />
      <ellipse cx="24" cy="23" rx="12.5" ry="11.5" fill="var(--cow-white)" stroke="var(--cow-outline)" strokeWidth="1.8" />
      <path d="M27.5 13.2c4.6-1 8.2 1.6 8.4 5-.8 2-3.6 2.6-5.6 1.2-1.6-1.2-3.4-3.6-2.8-6.2z" fill="var(--cow-spot)" />
      <circle cx="19.5" cy="22" r="1.9" fill="var(--cow-pupil)" />
      <circle cx="28.5" cy="22" r="1.9" fill="var(--cow-pupil)" />
      <ellipse cx="24" cy="31" rx="8.6" ry="5.6" fill="var(--cow-muzzle)" stroke="var(--cow-outline)" strokeWidth="1.6" />
      <ellipse cx="21" cy="30" rx="1.1" ry="1.5" fill="var(--cow-muzzle-dark)" />
      <ellipse cx="27" cy="30" rx="1.1" ry="1.5" fill="var(--cow-muzzle-dark)" />
    </svg>
  )
}
