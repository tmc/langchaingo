package chains

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

type testConversationalRetriever struct{}

func (t testConversationalRetriever) GetRelevantDocuments(_ context.Context, query string) ([]schema.Document, error) { // nolint: lll
	if query == "What did the president say about Ketanji Brown Jackson" {
		return []schema.Document{
			// nolint: lll
			{
				PageContent: "Tonight. I call on the Senate to: Pass the Freedom to Vote Act. Pass the John Lewis Voting Rights Act. And while you’re at it, pass the Disclose Act so Americans can know who is funding our elections. \n\nTonight, I’d like to honor someone who has dedicated his life to serve this country: Justice Stephen Breyer—an Army veteran, Constitutional scholar, and retiring Justice of the United States Supreme Court. Justice Breyer, thank you for your service. \n\nOne of the most serious constitutional responsibilities a President has is nominating someone to serve on the United States Supreme Court. \n\nAnd I did that 4 days ago, when I nominated Circuit Court of Appeals Judge Ketanji Brown Jackson. One of our nation’s top legal minds, who will continue Justice Breyer’s legacy of excellence.",
			},
			// nolint: lll
			{
				PageContent: "A former top litigator in private practice. A former federal public defender. And from a family of public school educators and police officers. A consensus builder. Since she’s been nominated, she’s received a broad range of support—from the Fraternal Order of Police to former judges appointed by Democrats and Republicans. \n\nAnd if we are to advance liberty and justice, we need to secure the Border and fix the immigration system. \n\nWe can do both. At our border, we’ve installed new technology like cutting-edge scanners to better detect drug smuggling.  \n\nWe’ve set up joint patrols with Mexico and Guatemala to catch more human traffickers.  \n\nWe’re putting in place dedicated immigration judges so families fleeing persecution and violence can have their cases heard faster. \n\nWe’re securing commitments and supporting partners in South and Central America to host more refugees and secure their own borders.",
			},
			// nolint: lll
			{
				PageContent: "And for our LGBTQ+ Americans, let’s finally get the bipartisan Equality Act to my desk. The onslaught of state laws targeting transgender Americans and their families is wrong. \n\nAs I said last year, especially to our younger transgender Americans, I will always have your back as your President, so you can be yourself and reach your God-given potential. \n\nWhile it often appears that we never agree, that isn’t true. I signed 80 bipartisan bills into law last year. From preventing government shutdowns to protecting Asian-Americans from still-too-common hate crimes to reforming military justice. \n\nAnd soon, we’ll strengthen the Violence Against Women Act that I first wrote three decades ago. It is important for us to show the nation that we can come together and do big things. \n\nSo tonight I’m offering a Unity Agenda for the Nation. Four big things we can do together.  \n\nFirst, beat the opioid epidemic.",
			},
			// nolint: lll
			{
				PageContent: "Tonight, I’m announcing a crackdown on these companies overcharging American businesses and consumers. \n\nAnd as Wall Street firms take over more nursing homes, quality in those homes has gone down and costs have gone up.  \n\nThat ends on my watch. \n\nMedicare is going to set higher standards for nursing homes and make sure your loved ones get the care they deserve and expect. \n\nWe’ll also cut costs and keep the economy going strong by giving workers a fair shot, provide more training and apprenticeships, hire them based on their skills not degrees. \n\nLet’s pass the Paycheck Fairness Act and paid leave.  \n\nRaise the minimum wage to $15 an hour and extend the Child Tax Credit, so no one has to raise a family in poverty. \n\nLet’s increase Pell Grants and increase our historic support of HBCUs, and invest in what Jill—our First Lady who teaches full-time—calls America’s best-kept secret: community colleges.",
			},
		}, nil
	}

	return []schema.Document{
		// nolint: lll
		{
			PageContent: "Tonight. I call on the Senate to: Pass the Freedom to Vote Act. Pass the John Lewis Voting Rights Act. And while you’re at it, pass the Disclose Act so Americans can know who is funding our elections. \n\nTonight, I’d like to honor someone who has dedicated his life to serve this country: Justice Stephen Breyer—an Army veteran, Constitutional scholar, and retiring Justice of the United States Supreme Court. Justice Breyer, thank you for your service. \n\nOne of the most serious constitutional responsibilities a President has is nominating someone to serve on the United States Supreme Court. \n\nAnd I did that 4 days ago, when I nominated Circuit Court of Appeals Judge Ketanji Brown Jackson. One of our nation’s top legal minds, who will continue Justice Breyer’s legacy of excellence.",
		},
		// nolint: lll
		{
			PageContent: "A former top litigator in private practice. A former federal public defender. And from a family of public school educators and police officers. A consensus builder. Since she’s been nominated, she’s received a broad range of support—from the Fraternal Order of Police to former judges appointed by Democrats and Republicans. \\n\\nAnd if we are to advance liberty and justice, we need to secure the Border and fix the immigration system. \\n\\nWe can do both. At our border, we’ve installed new technology like cutting-edge scanners to better detect drug smuggling.  \\n\\nWe’ve set up joint patrols with Mexico and Guatemala to catch more human traffickers.  \\n\\nWe’re putting in place dedicated immigration judges so families fleeing persecution and violence can have their cases heard faster. \\n\\nWe’re securing commitments and supporting partners in South and Central America to host more refugees and secure their own borders.",
		},
		// nolint: lll
		{
			PageContent: "Madam Speaker, Madam Vice President, our First Lady and Second Gentleman. Members of Congress and the Cabinet. Justices of the Supreme Court. My fellow Americans.  \n\nLast year COVID-19 kept us apart. This year we are finally together again. \n\nTonight, we meet as Democrats Republicans and Independents. But most importantly as Americans. \n\nWith a duty to one another to the American people to the Constitution. \n\nAnd with an unwavering resolve that freedom will always triumph over tyranny. \n\nSix days ago, Russia’s Vladimir Putin sought to shake the foundations of the free world thinking he could make it bend to his menacing ways. But he badly miscalculated. \n\nHe thought he could roll into Ukraine and the world would roll over. Instead he met a wall of strength he never imagined. \n\nHe met the Ukrainian people. \n\nFrom President Zelenskyy to every Ukrainian, their fearlessness, their courage, their determination, inspires the world",
		},
		// nolint: lll
		{
			PageContent: "As Ohio Senator Sherrod Brown says, “It’s time to bury the label “Rust Belt.” \\n\\nIt’s time. \\n\\nBut with all the bright spots in our economy, record job growth and higher wages, too many families are struggling to keep up with the bills.  \\n\\nInflation is robbing them of the gains they might otherwise feel. \\n\\nI get it. That’s why my top priority is getting prices under control. \\n\\nLook, our economy roared back faster than most predicted, but the pandemic meant that businesses had a hard time hiring enough workers to keep up production in their factories. \\n\\nThe pandemic also disrupted global supply chains. \\n\\nWhen factories close, it takes longer to make goods and get them from the warehouse to the store, and prices go up. \\n\\nLook at cars. \\n\\nLast year, there weren’t enough semiconductors to make all the cars that people wanted to buy. \\n\\nAnd guess what, prices of automobiles went up. \\n\\nSo—we have a choice. \\n\\nOne way to fight inflation is to drive down wages and make Americans poorer.",
		},
	}, nil
}

var _ schema.Retriever = testConversationalRetriever{}

func TestConversationalRetrievalQA(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	ctx := context.Background()

	llm, err := openai.New()
	require.NoError(t, err)

	combinedStuffQAChain := LoadStuffQA(llm)
	combinedQuestionGeneratorChain := LoadCondenseQuestionGenerator(llm)
	r := testConversationalRetriever{}

	chain := NewConversationalRetrievalQA(
		combinedStuffQAChain,
		combinedQuestionGeneratorChain,
		r,
		memory.NewConversationBuffer(memory.WithReturnMessages(true)),
	)
	result, err := Run(ctx, chain, "What did the president say about Ketanji Brown Jackson")
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Ketanji Brown Jackson"), "expected Ketanji Brown Jackson in result")

	result, err = Run(ctx, chain, "Did he mention who she succeeded")
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Justice Stephen Breyer"), "expected  Justice Stephen Breyer in result")
}

func TestConversationalRetrievalQAInvalidMemoryValue(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	ctx := context.Background()

	llm, err := openai.New()
	require.NoError(t, err)

	combinedStuffQAChain := LoadStuffQA(llm)
	combinedQuestionGeneratorChain := LoadCondenseQuestionGenerator(llm)
	r := testConversationalRetriever{}

	chain := NewConversationalRetrievalQA(
		combinedStuffQAChain,
		combinedQuestionGeneratorChain,
		r,
		memory.NewConversationBuffer(memory.WithReturnMessages(false)),
	)
	_, err = Run(ctx, chain, "What did the president say about Ketanji Brown Jackson")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), ErrMemoryValuesWrongType.Error()), "expected valid error to be thrown")
}

func TestConversationalRetrievalQAFromLLM(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	ctx := context.Background()

	r := testConversationalRetriever{}
	llm, err := openai.New()
	require.NoError(t, err)

	chain := NewConversationalRetrievalQAFromLLM(llm, r, memory.NewConversationBuffer(memory.WithReturnMessages(true)))
	result, err := Run(context.Background(), chain, "What did the president say about Ketanji Brown Jackson")
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Ketanji Brown Jackson"), "expected Ketanji Brown Jackson in result")

	result, err = Run(ctx, chain, "Did he mention who she succeeded")
	require.NoError(t, err)
	require.True(t, strings.Contains(result, " Justice Stephen Breyer"), "expected  Justice Stephen Breyer in result")
}

func TestConversationalRetrievalQAFromLLMWithConversationTokenBuffer(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	ctx := context.Background()

	r := testConversationalRetriever{}
	llm, err := openai.New()
	require.NoError(t, err)

	chain := NewConversationalRetrievalQAFromLLM(
		llm,
		r,
		memory.NewConversationTokenBuffer(llm, 2000, memory.WithReturnMessages(true)),
	)
	result, err := Run(context.Background(), chain, "What did the president say about Ketanji Brown Jackson")
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Ketanji Brown Jackson"), "expected Ketanji Brown Jackson in result")

	result, err = Run(ctx, chain, "Did he mention who she succeeded")
	require.NoError(t, err)
	require.True(t, strings.Contains(result, " Justice Stephen Breyer"), "expected  Justice Stephen Breyer in result")
}
