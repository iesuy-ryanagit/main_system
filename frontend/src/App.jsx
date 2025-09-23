import React, { useState, useEffect } from 'react';
import './App.css';

const API_URL = 'http://192.168.49.2:30080'; // ← your Minikube API

function App() {
  const [form, setForm] = useState({
    date: '',
    location: '',
    item: '',
    amount: ''
  });

  const [spendings, setSpendings] = useState([]); // Default to an empty array
  const [summary, setSummary] = useState([]); // Default to an empty array
  const [message, setMessage] = useState('');

  // Fetch spendings
  const fetchSpendings = async () => {
    try {
      const res = await fetch(`${API_URL}/spending/`);
      const data = await res.json();
      if (Array.isArray(data)) {
        setSpendings(data);  // Ensure data is an array
      } else {
        console.error('Invalid data format for spendings:', data);
      }
    } catch (err) {
      console.error('Failed to fetch spendings', err);
    }
  };

  // Fetch monthly summary
  const fetchSummary = async () => {
    try {
      const res = await fetch(`${API_URL}/summary/monthly`);
      const data = await res.json();
      if (Array.isArray(data)) {
        setSummary(data);  // Ensure data is an array
      } else {
        console.error('Invalid data format for summary:', data);
      }
    } catch (err) {
      console.error('Failed to fetch summary', err);
    }
  };

  useEffect(() => {
    fetchSpendings();
    fetchSummary(); // Fetch summary on initial load
  }, []);

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setMessage('');
    try {
      const res = await fetch(`${API_URL}/spending/`, { // Add trailing slash
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          date: form.date,
          location: form.location,
          item: form.item,
          amount: Number(form.amount),
        }),
      });
      if (res.ok) {
        setMessage('登録完了！');
        setForm({ date: '', location: '', item: '', amount: '' });
        fetchSpendings(); // Refresh spending data
        fetchSummary();   // Refresh summary data
      } else {
        const text = await res.text();
        setMessage(`エラー: ${text}`);
      }
    } catch (err) {
      setMessage('通信エラー');
    }
  };

  // Delete spending entry
  const handleDelete = async (id) => {
    if (window.confirm('本当に削除しますか？')) {
      try {
        const res = await fetch(`${API_URL}/spending/${id}`, {
          method: 'DELETE',
        });

        if (res.ok) {
          setMessage('削除完了！');
          fetchSpendings();
          fetchSummary(); // Refresh summary after deletion
        } else {
          const text = await res.text();
          setMessage(`削除エラー: ${text}`);
        }
      } catch (err) {
        setMessage('通信エラー');
      }
    }
  };

  return (
    <div style={{ maxWidth: 500, margin: 'auto', padding: 20 }}>
      <h1>Expenditure Management System</h1>
      <form onSubmit={handleSubmit}>
        <div>
          <label>日付: </label>
          <input type="date" name="date" value={form.date} onChange={handleChange} required />
        </div>
        <div>
          <label>場所: </label>
          <input type="text" name="location" value={form.location} onChange={handleChange} required />
        </div>
        <div>
          <label>品目: </label>
          <input type="text" name="item" value={form.item} onChange={handleChange} required />
        </div>
        <div>
          <label>金額: </label>
          <input type="number" name="amount" value={form.amount} onChange={handleChange} required min="0" />
        </div>
        <button type="submit" style={{ marginTop: 10 }}>登録</button>
      </form>
      {message && <p>{message}</p>}

      <h2>支出一覧</h2>
      <ul>
        {spendings && spendings.length > 0 ? (
          spendings.map((s) => (
            <li key={s.id}>
              {s.date} | {s.location} | {s.item} | ¥{s.amount}
              <button
                onClick={() => handleDelete(s.id)}
                style={{ marginLeft: 10, color: 'red' }}
              >
                削除
              </button>
            </li>
          ))
        ) : (
          <p>No spendings found</p>
        )}
      </ul>

      <h2>月別サマリー</h2>
      <ul>
        {summary && summary.length > 0 ? (
          summary.map((s) => (
            <li key={s.month}>
              {s.month} 合計: ¥{s.total}
            </li>
          ))
        ) : (
          <p>No summary data available</p>
        )}
      </ul>
    </div>
  );
}

export default App;
