export default function Sidebar({ user, menu, aktif, onSelect, onLogout }) {
  return (
    <aside className="side">
      <div className="brand">Rejeki Baru Farm<small>Pembukuan peternakan</small></div>
      {menu.map(g => (
        <div key={g.grup} style={{ display: 'contents' }}>
          <div className="grp">{g.grup}</div>
          {g.items.map(([k, label]) => (
            <button key={k} className={'nav' + (aktif === k ? ' on' : '')} onClick={() => onSelect(k)}>{label}</button>
          ))}
        </div>
      ))}
      <div className="who"><b>{user.username}</b>{user.role}
        <div style={{ marginTop: 8 }}><button className="ghost" onClick={onLogout}>Keluar</button></div>
      </div>
    </aside>
  )
}
