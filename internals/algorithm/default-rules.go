package algorithm

// DefaultPopularMenuItemIDs is the cold-start fallback menu order.
// It is also used to pad DB popular items when historical data is sparse.
var DefaultPopularMenuItemIDs = []string{
	"1e58b700-9091-4c9e-a9ac-6c4550fea44e", // Chicken biryani
	"42e73143-c837-45e3-aba0-b66cdaed317e", // Mutton biryani
	"4b4d1adb-7621-41d7-a1a4-46662215643c", // Chicken Thali
	"4a7c2a29-a91a-48bf-b3d8-dc9487cc02a1", // Mutton Thali
	"570d540d-9946-4ae4-a7f4-56140b303ea8", // Fish Thali
	"8b53258b-2ab3-42e9-8390-07cd86e6b163", // Veg. Thali
	"2fcfcd7e-dd98-40ee-a753-ae1e44649805", // Steam Buff Momo
	"e5accadf-1e1c-4eaf-ac60-c6239d96a61a", // Steam Chicken Momo
	"95345950-2bfa-43e7-8f42-21d02eeba664", // Fried Chicken Momo
	"3872d72a-2fbb-4a1b-88d1-4685362e053b", // Veg Momo
	"a25caecf-bddf-4ce5-a5b2-8b4c088f79af", // Chicken fried rice
	"d59ff4d3-012a-4936-9b40-ba15772c1dcb", // Egg fried rice
	"68fc1479-96b7-4c6c-8b67-2cf27abb7a00", // Veg fried rice
	"9247683d-5ae1-4820-9142-11f9ab0be559", // Chicken pizza
	"d0f00cfd-57f8-4e1c-8368-4c26cb699938", // Cheese pizza
	"d01139a6-d745-4366-9e6a-16447b597076", // Mushroom pizza
	"70afd73c-7258-40f0-874a-b95fdd459bb9", // Chicken burger
	"9852fe5e-e9fc-4b48-b339-f0744aa4d9e3", // Veg burger
	"cbfe5ba6-bd59-4a23-a4de-3339bf9a0247", // Sprite
	"147a52d7-e468-4d0e-a232-f1d9f467475e", // Pepsi
	"0db94cc7-8f83-4664-9a19-638b2dc9316f", // Fanta
	"1fd2f965-1b27-4a0d-b608-13974777b247", // Coke (250ml)
	"654857c7-adf1-4915-889d-79064b1860bc", // Mountain dew
	"7b6e4e31-e5fd-4e15-853e-9cded85d6c2e", // Xtreme
	"64cf70a8-b6e7-4c07-830b-358c6365d6ad", // Rasgulla
	"8fe14898-0db1-4980-8f6e-51fd42b21a8d", // Laddu
	"989383ea-547f-4958-b0aa-22c787c2942c", // Pastry
	"a8ba0c41-f7e2-4d00-9f4e-26569ef6881e", // Ice-cream
}

// DefaultAssociationRules returns deterministic "if-then" rules for cold start.
// These are used only when Apriori cannot produce any rule from real orders.
func DefaultAssociationRules() []AssociationRule {
	addRules := func(bucket []AssociationRule, antecedent string, consequents ...string) []AssociationRule {
		for i, c := range consequents {
			bucket = append(bucket, AssociationRule{
				Antecedent: ItemSet{antecedent},
				Consequent: ItemSet{c},
				Support:    0.30 - float64(i)*0.01,
				Confidence: 0.75 - float64(i)*0.02,
				Lift:       1.60 - float64(i)*0.05,
			})
		}
		return bucket
	}

	var rules []AssociationRule

	// Biryani combos
	rules = addRules(rules, "1e58b700-9091-4c9e-a9ac-6c4550fea44e",
		"147a52d7-e468-4d0e-a232-f1d9f467475e", // Pepsi
		"cbfe5ba6-bd59-4a23-a4de-3339bf9a0247", // Sprite
		"a8ba0c41-f7e2-4d00-9f4e-26569ef6881e", // Ice-cream
	)
	rules = addRules(rules, "42e73143-c837-45e3-aba0-b66cdaed317e",
		"1fd2f965-1b27-4a0d-b608-13974777b247", // Coke
		"654857c7-adf1-4915-889d-79064b1860bc", // Mountain dew
		"64cf70a8-b6e7-4c07-830b-358c6365d6ad", // Rasgulla
	)

	// Momo combos
	rules = addRules(rules, "2fcfcd7e-dd98-40ee-a753-ae1e44649805",
		"cbfe5ba6-bd59-4a23-a4de-3339bf9a0247", // Sprite
		"654857c7-adf1-4915-889d-79064b1860bc", // Mountain dew
		"a8ba0c41-f7e2-4d00-9f4e-26569ef6881e", // Ice-cream
	)
	rules = addRules(rules, "e5accadf-1e1c-4eaf-ac60-c6239d96a61a",
		"147a52d7-e468-4d0e-a232-f1d9f467475e", // Pepsi
		"0db94cc7-8f83-4664-9a19-638b2dc9316f", // Fanta
		"989383ea-547f-4958-b0aa-22c787c2942c", // Pastry
	)
	rules = addRules(rules, "3872d72a-2fbb-4a1b-88d1-4685362e053b",
		"cbfe5ba6-bd59-4a23-a4de-3339bf9a0247", // Sprite
		"0db94cc7-8f83-4664-9a19-638b2dc9316f", // Fanta
		"8fe14898-0db1-4980-8f6e-51fd42b21a8d", // Laddu
	)

	// Rice + pizza + burger combos
	rules = addRules(rules, "a25caecf-bddf-4ce5-a5b2-8b4c088f79af",
		"654857c7-adf1-4915-889d-79064b1860bc", // Mountain dew
		"a8ba0c41-f7e2-4d00-9f4e-26569ef6881e", // Ice-cream
		"989383ea-547f-4958-b0aa-22c787c2942c", // Pastry
	)
	rules = addRules(rules, "9247683d-5ae1-4820-9142-11f9ab0be559",
		"1fd2f965-1b27-4a0d-b608-13974777b247", // Coke
		"cbfe5ba6-bd59-4a23-a4de-3339bf9a0247", // Sprite
		"a8ba0c41-f7e2-4d00-9f4e-26569ef6881e", // Ice-cream
	)
	rules = addRules(rules, "70afd73c-7258-40f0-874a-b95fdd459bb9",
		"147a52d7-e468-4d0e-a232-f1d9f467475e", // Pepsi
		"1fd2f965-1b27-4a0d-b608-13974777b247", // Coke
		"a8ba0c41-f7e2-4d00-9f4e-26569ef6881e", // Ice-cream
	)
	rules = addRules(rules, "9852fe5e-e9fc-4b48-b339-f0744aa4d9e3",
		"0db94cc7-8f83-4664-9a19-638b2dc9316f", // Fanta
		"cbfe5ba6-bd59-4a23-a4de-3339bf9a0247", // Sprite
		"8fe14898-0db1-4980-8f6e-51fd42b21a8d", // Laddu
	)

	// Thali combos
	rules = addRules(rules, "4b4d1adb-7621-41d7-a1a4-46662215643c",
		"147a52d7-e468-4d0e-a232-f1d9f467475e", // Pepsi
		"64cf70a8-b6e7-4c07-830b-358c6365d6ad", // Rasgulla
		"989383ea-547f-4958-b0aa-22c787c2942c", // Pastry
	)
	rules = addRules(rules, "8b53258b-2ab3-42e9-8390-07cd86e6b163",
		"cbfe5ba6-bd59-4a23-a4de-3339bf9a0247", // Sprite
		"8fe14898-0db1-4980-8f6e-51fd42b21a8d", // Laddu
		"a8ba0c41-f7e2-4d00-9f4e-26569ef6881e", // Ice-cream
	)

	return rules
}
