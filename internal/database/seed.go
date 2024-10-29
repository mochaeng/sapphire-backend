package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

func Seed(store *store.Store, db *sql.DB) {
	ctx := context.Background()

	tx, _ := db.BeginTx(ctx, nil)
	users := generateUsers(100)
	for _, user := range users {
		if err := store.User.Create(ctx, tx, user); err != nil {
			log.Panicf("error creating user: [%+v]: %s", user, err)
		}
	}
	tx.Commit()

	posts := generatePosts(200, users)
	for _, post := range posts {
		if err := store.Post.Create(ctx, post); err != nil {
			log.Panicf("error creating post: %s", err)
		}
	}

	comments := generateComments(200, users, posts)
	for _, comment := range comments {
		if err := store.Comment.Create(ctx, comment); err != nil {
			log.Panicf("error creating comment: %s", err)
		}
	}

	followers := generateFollowers(100, users)
	for _, follower := range followers {
		if err := store.User.Follow(ctx, follower.FollowerID, follower.FollowedID); err != nil {
			log.Panicf("error following user: %s", err)
		}
	}

	log.Println("finishing seeding database")
	return
}

func generateUsers(amount int) []*models.User {
	users := make([]*models.User, amount)
	for i := 0; i < amount; i++ {
		username := usernames[i%len(usernames)] + fmt.Sprintf("_%d", i)
		users[i] = &models.User{
			Username:  username,
			FirstName: names[rand.Intn(len(names))],
			LastName:  names[rand.Intn(len(names))],
			Email:     username + fmt.Sprintf("@gmail.com"),
			Role: models.Role{
				ID: 1,
			},
		}
	}
	return users
}

func generatePosts(amount int, users []*models.User) []*models.Post {
	posts := make([]*models.Post, amount)
	for i := 0; i < amount; i++ {
		user := users[rand.Intn(len(users))]
		posts[i] = &models.Post{
			User:    user,
			Tittle:  tittles[rand.Intn(len(tittles))],
			Content: contents[rand.Intn(len(contents))],
			Tags: []string{
				tags[rand.Intn(len(tags))],
				tags[rand.Intn(len(tags))],
			},
		}
	}
	return posts
}

func generateComments(amount int, users []*models.User, posts []*models.Post) []*models.Comment {
	comments := make([]*models.Comment, amount)
	for i := 0; i < amount; i++ {
		user := users[rand.Intn(len(users))]
		post := posts[rand.Intn(len(posts))]
		comments[i] = &models.Comment{
			UserId:  user.ID,
			PostId:  post.ID,
			Content: commentaries[rand.Intn(len(commentaries))],
		}
	}
	return comments
}

func generateFollowers(amount int, users []*models.User) []*models.Follower {
	followers := []*models.Follower{}
	seen := make(map[string]bool)
	for i := 0; i < amount; i++ {
		user1 := users[rand.Intn(len(users))]
		user2 := users[rand.Intn(len(users))]
		if user1.ID != user2.ID {
			key := fmt.Sprintf("%d-%d", user1.ID, user2.ID)
			if _, exists := seen[key]; !exists {
				follower := models.Follower{
					FollowerID: user1.ID,
					FollowedID: user2.ID,
				}
				followers = append(followers, &follower)
				seen[key] = true
			}
		}
	}
	return followers
}

var usernames = []string{
	"TechWhiz101", "SkyGazer88", "EpicExplorer", "PixelPainter", "VibrantVibes",
	"DigitalNomad", "StarryNight", "SilentMuse", "WiseOwl", "TheCuriousCat",
	"ByteWarrior", "PixelPioneer", "DreamChaser", "MellowSun", "QuestSeeker",
	"ByteBender", "CodeNinja", "EchoEcho", "OceanMystic", "BookWormVibes",
	"ArtsySoul", "ZenGamer", "LivelyLion", "CloudDweller", "Wanderlust2024",
	"PandaCoder", "CityLights", "QuantumDreamer", "EpicScribe", "NebulaNova",
	"AstralArtist", "LaughingTiger", "MysticMeadow", "SunsetChaser", "CodeEagle",
	"CuriousOtter", "BoldButterfly", "CelestialSky", "RetroRider", "LazyCoffee",
}

var names = []string{
	"Liam", "Olivia", "Noah", "Emma", "Ava", "Elijah", "Sophia", "James",
	"Amelia", "Isabella", "Mason", "Mia", "Ethan", "Luna", "Lucas", "Charlotte",
	"Benjamin", "Scarlett", "Henry", "Aurora", "Ella", "Jackson", "Harper",
	"Oliver", "Grace", "Mateo", "Nova", "Aiden", "Layla", "Leo", "Riley", "Evelyn",
	"Samuel", "Zoe", "Carter", "Hazel", "Alexander", "Lily", "Sebastian", "Penelope",
}

var tittles = []string{
	"Exploring the Future of AI", "Top 10 Travel Destinations for 2024", "How to Stay Productive Remotely",
	"The Best Books I've Read This Year", "Why Minimalism Changed My Life", "Tips for a Healthy Morning Routine",
	"Understanding Cryptocurrency", "Hidden Gems for Movie Lovers", "A Beginner's Guide to Yoga",
	"What I Learned from Starting a Business", "Top Coding Tips for Beginners", "Secrets to a Happy Life",
	"Best Cafés to Visit in NYC", "How to Boost Your Creativity", "The Benefits of Meditation",
	"A Guide to Sustainable Living", "My Journey in Learning to Code", "Is Social Media Doing More Harm Than Good?",
	"10 Ways to Reduce Stress", "Exploring the World of Photography", "Why I Switched to a Plant-Based Diet",
	"Learning a New Language: Tips and Tricks", "Understanding Climate Change", "Best Fitness Apps of the Year",
	"How to Save Money Effectively", "A Day in the Life of a Freelancer", "The Ultimate Road Trip Playlist",
	"Why We Should Protect Wildlife", "Exploring Ancient Civilizations", "How to Improve Your Mental Health",
	"Top Skills for the Future Job Market", "Minimalist Wardrobe Essentials", "Is AI Going to Take Over?",
	"A Guide to Mindfulness", "Best Apps for Staying Organized", "The Art of Journaling",
	"The Importance of Voting", "Tips for Solo Travelers", "Why Reading is a Superpower",
	"Exploring the World of Indie Games",
}

var tags = []string{
	"Technology", "Travel", "Lifestyle", "Health", "Productivity",
	"Food", "Nature", "Art", "Photography", "Fitness",
	"SelfImprovement", "Education", "Music", "Movies", "Business",
	"Coding", "Environment", "Finance", "MentalHealth", "Books",
}

var contents = []string{
	"Excited to start my journey in coding! Any tips for beginners?",
	"Just watched an incredible documentary on the universe. Mind blown!",
	"Coffee and books – my favorite combo on a rainy day.",
	"Trying to figure out the best way to stay organized. Any app suggestions?",
	"The mountains are calling, and I must go!",
	"Couldn’t resist sharing this cute photo of my pet today!",
	"Exploring the world one day at a time. Travel is the best education!",
	"Here’s why I think minimalism changed my life for the better.",
	"Learning to let go of what doesn’t serve me. Feels freeing.",
	"Finished an amazing novel last night. Book lovers, where you at?",
	"Why isn’t there more focus on mental health in schools?",
	"Feeling inspired after a long hike. Nature has all the answers.",
	"Cooking my way through different cuisines. Any recipe recs?",
	"Just started yoga and already feel the benefits!",
	"Trying to stay productive while working from home – send help!",
	"What podcasts are you all listening to lately?",
	"The power of a good playlist – can change your whole day!",
	"Every day is a new chance to grow and learn.",
	"Celebrating small wins because they’re just as important!",
	"Exploring meditation as a way to stay grounded. Loving it!",
	"Just completed my first freelance project! Onwards and upwards.",
	"Been thinking a lot about sustainable living. Where to start?",
	"A warm cup of tea and a good book – simple pleasures.",
	"Taking it one day at a time and appreciating every moment.",
	"Any favorite movie recommendations? I need a new obsession.",
	"Today I learned: everything we need is within us.",
	"Diving into photography and loving the creative process.",
	"Why is traveling solo so empowering? Must try it once!",
	"Currently obsessed with learning about ancient civilizations.",
	"Trying out new fitness apps – which ones do you recommend?",
	"Saving money hacks: share yours!",
	"Couldn’t be happier about the little things today.",
	"Who else loves a good sunset?",
	"Just saw a beautiful sunrise – perfect start to the day!",
	"Taking up a new hobby this month. Any suggestions?",
	"The best part of my day: unwinding with some soft music.",
	"Made a healthy dinner tonight – actually tasted amazing!",
	"How do you balance work and personal projects?",
	"Life lesson: it’s okay to not have everything figured out.",
	"Every small step adds up. Keep going.",
	"Self-care is not a luxury; it’s a necessity.",
}

var commentaries = []string{
	"Totally agree with this!", "Great point, thanks for sharing!", "Couldn't have said it better myself!",
	"Love this perspective!", "This really resonated with me.", "Interesting, I'll have to think about this.",
	"Thanks for the tips!", "Beautifully put!", "I needed this reminder today.", "Completely agree!",
	"Couldn't stop nodding while reading this!", "Well said!", "So true!", "Couldn't agree more!",
	"This is exactly what I was looking for!", "Thanks for the inspiration!", "This made my day!",
	"Super helpful, thank you!", "I love your content!", "Wow, this is amazing!",
	"I was just thinking about this!", "Totally relatable.", "This is so encouraging!",
	"Such a unique take!", "I’ve been trying to do the same!", "Incredible insight!", "I wish more people thought this way.",
	"Thank you for sharing this!", "Spot on!", "I feel the same way!", "Definitely going to try this out!",
	"Such a great idea!", "This hit home.", "I couldn’t agree more.", "Really made me think!",
	"Thanks for breaking this down so well.", "This just changed my perspective.", "Perfect timing on this post!",
	"This deserves more attention!", "Yes, yes, yes! Love this!",
}
