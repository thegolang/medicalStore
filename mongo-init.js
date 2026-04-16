// Initialize medicalstore database with default admin user
db = db.getSiblingDB('medicalstore');

// Create collections
db.createCollection('users');
db.createCollection('medicines');
db.createCollection('categories');
db.createCollection('suppliers');
db.createCollection('sales');
db.createCollection('purchases');

// Default admin user (password: admin123 - bcrypt hashed)
db.users.insertOne({
  username: 'admin',
  email: 'admin@medicalstore.com',
  // bcrypt hash of "admin123" (cost=10)
  password: '$2a$10$pQUaGo4esLMP6IjpraPwjeIFX6QUi9x57IYU.xnIooAmy15/vqjWi',
  role: 'admin',
  full_name: 'Store Administrator',
  active: true,
  created_at: new Date(),
  updated_at: new Date()
});

// Default categories
db.categories.insertMany([
  { name: 'Antibiotics', description: 'Medicines that kill or inhibit bacteria', active: true, created_at: new Date() },
  { name: 'Analgesics', description: 'Pain relief medicines', active: true, created_at: new Date() },
  { name: 'Antipyretics', description: 'Fever reducing medicines', active: true, created_at: new Date() },
  { name: 'Antacids', description: 'Medicines for acid reflux and indigestion', active: true, created_at: new Date() },
  { name: 'Vitamins & Supplements', description: 'Nutritional supplements', active: true, created_at: new Date() },
  { name: 'Antidiabetics', description: 'Medicines for diabetes management', active: true, created_at: new Date() },
  { name: 'Cardiovascular', description: 'Heart and blood pressure medicines', active: true, created_at: new Date() },
  { name: 'Dermatology', description: 'Skin care medicines', active: true, created_at: new Date() }
]);

print('Database initialized successfully!');
