import matplotlib.pyplot as plt
import pandas as pd

columns = ["course", "trade", "dex_name", "blockNumber"]
df = pd.read_csv("~/swap-listener/static/price.csv", names = columns, header=None)
uni_buys = []
sushi_buys = []
uni_sells = []
sushi_sells = []
for i in df.blockNumber.unique():
    block = df.loc[df.blockNumber == i]
    uni = block.loc[block.dex_name == "Uniswap"]
    sushi = block.loc[block.dex_name == "Sushiswap"]
    print(uni.loc(uni.trade == "Buy").mean())
    uni_buys.append(uni.loc(uni.trade == "Buy").mean()[0])
    uni_sells.append(uni.loc(uni.trade == "Sell").mean()[0])
    sushi_buys.append(sushi.loc(sushi.trade == "Buy").mean()[0])
    sushi_sells.append(sushi.loc(sushi.trade == "Sell").mean()[0])
plt.plot(uni_buys)
plt.plot(sushi_buys)
plt.plot(uni_sells)
plt.plot(sushi_sells)
# plt.ylabel('some numbers')
plt.show()
plt.savefig("./graph.png")