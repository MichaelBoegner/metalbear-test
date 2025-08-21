import { useEffect, useState } from "react";
import "./App.css"; // import the stylesheet

const colors = ["#549", "#18d", "#d31", "#2a4", "#db1"];

export default function App() {
  const [entries, setEntries] = useState([]);
  const [entryValue, setEntryValue] = useState("");
  const [color, setColor] = useState(colors[Math.floor(Math.random() * colors.length)]);
  const [hostAddress, setHostAddress] = useState("");

  // Fetch guestbook entries
  const fetchEntries = async () => {
    try {
      const res = await fetch("/lrange/guestbook");
      const data = await res.json();
      setEntries(data);
    } catch (err) {
      console.error("Failed to fetch entries", err);
    }
  };

  // Submit new entry
  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!entryValue.trim()) return;

    setEntries([...entries, "..."]); // placeholder
    try {
      const res = await fetch(`/rpush/guestbook/${encodeURIComponent(entryValue)}`);
      const data = await res.json();
      setEntries(data);
      setEntryValue("");
    } catch (err) {
      console.error("Failed to add entry", err);
    }
  };

  useEffect(() => {
    setHostAddress(window.location.href);
    fetchEntries();
    const interval = setInterval(fetchEntries, 1000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div style={{ "--dynamic-color": color }}>
      <div id="header">
        <h1 className="header-title">mirrord Guestbook</h1>
      </div>

      <div id="guestbook-entries">
        {entries.length === 0 ? (
          <p>Waiting for database connection...</p>
        ) : (
          entries.map((val, idx) => <p key={idx}>{val}</p>)
        )}
      </div>

      <div>
        <form id="guestbook-form" onSubmit={handleSubmit}>
          <input
            id="guestbook-entry-content"
            type="text"
            value={entryValue}
            onChange={(e) => setEntryValue(e.target.value)}
          />
          <button id="guestbook-submit" type="submit">
            Submit
          </button>
        </form>
      </div>

      <div>
        <h2 id="guestbook-host-address">{hostAddress}</h2>
        <p>
          <a href="/env">/env</a> <a href="/info">/info</a>
        </p>
      </div>
    </div>
  );
}
