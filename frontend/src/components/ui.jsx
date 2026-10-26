export const Field = ({ label, children }) => <label className="f">{label}{children}</label>

export const Stat = ({ label, value, hi }) => <div className={'stat' + (hi ? ' hi' : '')}><span>{label}</span><b>{value}</b></div>

export const Empty = ({ text }) => <div className="empty">{text}</div>
