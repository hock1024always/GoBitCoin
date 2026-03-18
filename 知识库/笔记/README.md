#  学习路线图

首先，欢迎大家查看我们的资料库，当你踏入区块链的领域之前，不妨先看看我们给出的学习路线吧

下面是登链社区给出的区块链开发者的学习路线图

![1e9526375eaf48d9a2eaa3a2a68d0e99.jpeg](https://i-blog.csdnimg.cn/direct/1e9526375eaf48d9a2eaa3a2a68d0e99.jpeg)

# 学习路线建议

对于一个区块链方向的学习者而言，首先要了解的是区块链理论知识，当你了解了区块链的理论知识之后，下面有三个方向来学习，可以通俗的理解为区块链方向的后端，前端以及技术应用。

![543506c8afc24947a771e99efe133268.png](https://i-blog.csdnimg.cn/direct/543506c8afc24947a771e99efe133268.png)



## **区块链理论知识**

这里，我们推荐两门课程：

### **北京大学肖臻老师的区块链公开课：**

[北京大学肖臻老师《区块链技术与应用》公开课_哔哩哔哩_bilibili](https://www.bilibili.com/video/BV1Vt411X7JF?t=0.0)

这门课程讲述了比特币，以太坊（PoS之前）的知识，非常详细，能为你的区块链学习开发打下坚实的基础！

### **Tintinland区块链基础通识课：**

[第一课-1：什么是区块链｜区块链通识基础_哔哩哔哩_bilibili](https://www.bilibili.com/video/BV1WS411A7h9?t=0.3)

这门课程主要涉及的有以太坊（PoS之后），包含PoS机制，标准区块链范式的内容。

## 区块链底层

首先，区块链的底层现在比较主流的语言有Golang和Rust，Rust链开发我们方向的@黎俊奕 大佬比较了解，随后请他在下面进行补充，这里简要介绍一下Go语言的三个方向：

### Go语言实现比特币

这是一个基础的BTC区块链构建，一个文件夹到底，对新手友好！

[Go语言实现比特币-完整教程-代码视频_哔哩哔哩_bilibili](https://www.bilibili.com/video/BV1EY4y1c7Yq?t=0.0)

### Go语言实现Web3

相比上面的视频，这个视频的内容要更加深入一点，涉及EVM虚拟机，加密货币等知识，相对深入一点

[【区块链开发】使用 Golang 从零开始构建区块链（34小时超详细课程）_哔哩哔哩_bilibili](https://www.bilibili.com/video/BV1hz42167qE?t=0.0)

### Go语言实现Web3应用底层

这个是使用go语言搭建一个Web3项目，实现了这个项目，你就具备了参加黑客松比赛的简单能力。

[【Web3 游戏开发】使用 Golang 构建去一个中心化扑克游戏（GameFi赛道）_哔哩哔哩_bilibili](https://www.bilibili.com/video/BV1ji421C7Kx?t=0.0)

## Web3开发

Web3，常见的开发方向就是基于某条区块链的应用开发（dApp）。而基于这种应用开发，类比前后端方向也有如下的三个方向：智能合约（区块链上的一种自动化程序），全栈应用开发，合约安全审计。

### Solidity开发

语法

喜欢文档的同学可以参考WFT学院的solidity三部分

 [Solidity入门 | WTF Academy](https://www.wtf.academy/docs/solidity-101/)

这个文档有配套的视频（[五里墩茶社](https://space.bilibili.com/615957867)大佬的学习笔记）

[[跟我学Solidity\] 第1日: HelloWeb3和数值类型_哔哩哔哩_bilibili](https://www.bilibili.com/video/BV1CP4y1z7pF?t=1.1)

还有就是崔棉大师的Solidity课程，业界好评

[Solidity8.0全面精通-01-Solidity8.0新特性_哔哩哔哩_bilibili](https://www.bilibili.com/video/BV1oZ4y1B7WS?t=0.2)

框架

这个是Patrick Collins大神最新的foundry教学课程，foundry和hardhat是Solisity开发最常用的两个框架，这里面也涉及了很多的Web3基础知识，值得一看

https://www.youtube.com/watch?v=-1GB6m39-rM

### 智能合约安全审计

合约审计的漏洞查找是一项非常高难度的工作，不同于一般程序员的常规查错，合约在部署之后是不可篡改的，从theDao造成的的以太坊分裂，到现在的动态攻击，合约安全审计一致是区块链上最终要的工作之一，值得一提的是我们工作室的一位毕业学长目前就在做合约审计的相关工作。下面这是Patrick Collins大神最新的安全审计课程，当你有了一定的合约开发基础，不妨来试一试安全审计吧。

https://www.youtube.com/watch?v=Y3WMkl0AFJk

### Dapp开发

如果你想要参加区块链方向最令人激动的黑客松比赛，那么Dapp开发毫无疑问是最常规的赛道，这样的开发分支在建议路线图中写了

## 区块链科学

区块链科学部分，我们主要聚焦区块链的扩容问题，想要了解这部分知识，可以先学习一下区块链Layer2是什么，下面是Web3实验室的视频，非常通俗！

暂时无法在成电飞书文档外展示此内容

SideChain

sidechain就是借助主链数据，使用自己的EVM进行交易的处理，最后将其拥有的数据同步给主链。

Rollup

Rollup 本质上是一条独立的区块链，拥有自己的虚拟机，不同于以太坊的区块是由多数节点认可来实现其合法性的，监控 Rollup 状态的一方可以将 “断言” 发送至以太坊，来说明交易是如何处理的；以太坊将决定是否接受这个断言，无论这个断言是否获得了 Rollup 上多数参与方的支持。而接受断言的选择就出现了两个分支：zkRollup和OptimismRollup

零知识证明理论：

https://www.bilibili.com/video/BV1N7xMeREEf/?spm_id_from=333.337.search-card.all.click

# 学习资源

## 社群

1. lllu_23(up主)：我的Web3方向的启蒙老师。搬运了很多油管的课程，建议观看他搬运的知识性视频，非常优秀。缺点是很多项目是在linux上面开发的，如果不是很熟悉linux的话很多大型项目会出很多bug；其二是由于Web3行业的时效性，很多比较老的视频是存在版本误差的，往往在学习coding很早的视频时会遇到很多问题。
2. TinTinLand：一个新兴的Web3社群，组织的公开课和有项目奖金的黑客松比赛还是很赞的，他们提供的区块链内容更偏向实用性，推荐的链和框架往往也都是最新的，值得关注！
3. 登链社区：侧重于技术性的文章，我心目中国内一等权威的区块链技术网站。
4. Chainlink预言机：经常能够听到一句话：“一个没有听说过Chainlink的区块链开发者一定不是一个优秀的开发者”，这句话可能有点夸大，但是Chainlink社区确实是一个国内外都很知名的社区，虽然我没有怎么看过他们的作品。
5. Rebase：一个侧重区块链底层技术原理的社区。

## 网站

WTF学院，一个由北大区块链技术社团开发的学习网站，在这里你可以学到Solidity语法，Ether.js，Langchain大模型，前端知识，以太坊虚拟机等内容。

https://www.wtf.academy/

生成数字藏品NFT的网站

https://promptbase.com/

下面的是一个各个链的基础知识以及黑客松比赛的网站

https://www.hackquest.io/zh

## 黑客松

黑客松是一种偏向于Web3全栈项目开发的短时性比赛，在黑客松比赛中得奖其一是有丰厚的奖金，其二是对于入职Web3公司有很大的加分项。常见的黑客松大致分为两类：

1. 第一种是奖金在5w$左右的大型黑客松，一般是需要成熟的项目，比较难得奖，适合资深开发者参与。
2. 第二种是小一些的黑客松，主要是某些链或者社区的推广项目，相对容易得奖，而且有些时候只要发布项目就可以获得一定数额的奖金，适合练手。
