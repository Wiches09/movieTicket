// mongo-init.js
db = db.getSiblingDB("booking_db");

// Seed Movies
// (Currently empty in running DB, using placeholders)
db.movies.insertMany([
  { _id: 1, title: "Interstellar", year: 2014 },
  { _id: 2, title: "Inception", year: 2010 },
  { _id: 3, title: "The Dark Knight", year: 2008 },
  { _id: 4, title: "The Batman", year: 2022 },
  { _id: 5, title: "Oppenheimer", year: 2023 },
  { _id: 6, title: "Dune: Part Two", year: 2024 },
]);

// Seed Admin User (Synced from running DB)
db.users.insertOne({
  _id: "k9mx3OUirAW6X3HE9SEWhTsS5v83",
  display_name: "admin",
  email: "admin@gmail.com",
  role: "admin",
  created_at: new Date(),
  updated_at: new Date(),
});

// Log success
print("Database 'booking_db' initialized with seed movies and admin user.");
