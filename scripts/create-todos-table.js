import pg from 'pg';

const { Client } = pg;

const client = new Client({
  host: '106.14.191.34',
  port: 5432,
  user: 'postgres',
  password: 'postgres',
  database: 'postgres'
});

async function createTable() {
  try {
    await client.connect();
    console.log('Connected to database');

    const createTableQuery = `
      CREATE TABLE IF NOT EXISTS public.todos (
        id SERIAL PRIMARY KEY,
        task TEXT,
        is_completed BOOLEAN DEFAULT false,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
      );
    `;

    await client.query(createTableQuery);
    console.log('Table "todos" created successfully');

    const result = await client.query('SELECT * FROM information_schema.tables WHERE table_name = $1', ['todos']);
    console.log('Table exists:', result.rows.length > 0);

    await client.end();
    console.log('Disconnected from database');
  } catch (error) {
    console.error('Error:', error.message);
    process.exit(1);
  }
}

createTable();