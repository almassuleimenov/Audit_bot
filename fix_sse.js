const fs = require('fs');
const file = 'audit-dashboard/src/app/DashboardClient.tsx';
let data = fs.readFileSync(file, 'utf8');
data = data.replace(
  "const eventSource = new EventSource('/api/stream/leads', {",
  "const API_URL = process.env.NEXT_PUBLIC_API_URL || 'https://audit-bot-1-j1xj.onrender.com';\n    const eventSource = new EventSource(`${API_URL}/api/stream/leads`, {"
);
fs.writeFileSync(file, data);
