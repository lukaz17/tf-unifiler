// Copyright (C) 2024  T-Force I/O
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package engine

import (
	"encoding/hex"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/tforce-io/tf-golib/stdx"
	"github.com/tforceaio/tf-unifiler/core"
	"github.com/tforceaio/tf-unifiler/db"
)

func TestFileSaveIResults(t *testing.T) {
	ctx := getTestDB("Main", "FileSaveIResults")
	exes, imgs := testingFileData()
	tests := []struct {
		group        string
		sessions     [][]*core.FileMultiHash
		collections  [][]string
		hashCount    int64
		mappingCount int64
		setCount     int64
		setHashCount int64
	}{
		{
			"separated_session",
			[][]*core.FileMultiHash{
				exes,
				imgs,
			},
			[][]string{
				{"exes"},
				{"imgs"},
			},
			24, 25, 2, 24,
		},
		{
			"multi_sessions",
			[][]*core.FileMultiHash{
				exes[0:3],
				exes[1:6],
			},
			[][]string{
				{"exes"},
				{"exes"},
			},
			6, 6, 1, 6,
		},
		{
			"multi_sets",
			[][]*core.FileMultiHash{
				exes[0:3],
				exes,
				imgs[1:9],
				imgs[8:],
			},
			[][]string{
				{"exes", "all"},
				{"exes"},
				{"imgs", "all"},
				{"imgs"},
			},
			23, 24, 3, 34,
		},
	}
	for _, tt := range tests {
		t.Run(tt.group, func(t *testing.T) {
			ctx.Reset()

			module := &MetadataModule{
				logger: log.Logger,
			}
			for j, ss := range tt.sessions {
				module.saveIResults(ctx, ss, false, tt.collections[j], "", "", false)
			}
			lastSession, err := ctx.GetLatestSession()
			if err != nil {
				t.Errorf("last session not found")
			}
			_, err = ctx.CountSessionChanges(lastSession.ID)
			if err != nil {
				t.Errorf("session changes cannot be retrieved")
			}
			hashCount, err := ctx.Count(&db.Hash{}, nil, nil)
			if err != nil || hashCount != tt.hashCount {
				t.Errorf("hash records count mismatch. expected %v actual %v", tt.hashCount, hashCount)
			}
			mappingCount, err := ctx.Count(&db.Mapping{}, nil, nil)
			if err != nil || mappingCount != tt.mappingCount {
				t.Errorf("mapping records count mismatch. expected %v actual %v", tt.mappingCount, mappingCount)
			}
			setCount, err := ctx.Count(&db.Set{}, nil, nil)
			if err != nil || setCount != tt.setCount {
				t.Errorf("set records count mismatch. expected %v actual %v", tt.setCount, setCount)
			}
			setHashCount, err := ctx.Count(&db.SetHash{}, nil, nil)
			if err != nil || setHashCount != tt.setHashCount {
				t.Errorf("set_hash records count mismatch. expected %v actual %v", tt.setHashCount, setHashCount)
			}
		})
	}
}

func bytesFromHex(hexStr string) stdx.Bytes {
	buf, _ := hex.DecodeString(hexStr)
	return stdx.Bytes(buf)
}

func createFileMultiHash(fileName string, size int64, md5, sha1, sha256, sha512 string) *core.FileMultiHash {
	return &core.FileMultiHash{
		Md5:      bytesFromHex(md5),
		Sha1:     bytesFromHex(sha1),
		Sha256:   bytesFromHex(sha256),
		Sha512:   bytesFromHex(sha512),
		Size:     size,
		FileName: fileName,
	}
}

func testingFileData() (exes, imgs []*core.FileMultiHash) {
	exes = []*core.FileMultiHash{
		createFileMultiHash("curl", 280872, "71192c83a60987b32e7c792a7d6c8d4a", "a40d55383afaeb0f838063daac0f83fa7a23cf0d", "c1a44dd5f1e5be612408eac67aed6f60fa6abdee6b3483955d40830649d03e26", "6e0317c730afb3cff14a40581f9a10f744def6e4234fa4adacb657219f8347873fc12331c36fe7e0d91488e1a731d69a38d205292d661be169103ffb55c5729e"),
		createFileMultiHash("gcc", 1026488, "9f999bff5baee68c86a218f85f3b2adb", "496fda1c4a1de4bc31ed34bab2e1587348fe7faa", "ddfc5a404fa3bbfc66ef65397d9db6269163bb5066ccc51ea469b6a477c640e5", "a1302556af3137c868e0e5dc602c4b3f7fef9122d363d81b668a37c8090a93bc7411f4be5740e8a40a46156640ca067a4dca4489e3aa796c59e7eab8e2c615e3"),
		createFileMultiHash("git", 3726520, "048f2515ebc535434d4df2b71a5a59bd", "c4070250597c287168f8e15c2f18d948b53dffca", "6bb5b0f75375c878cd0dda71ffdebe9adc664026f3e557fb151a71290e741dff", "180fed998e870137681a69e0f2f9a50450d565f99331710ec57af988faf885e45bdd1db0e2bb37d9b3af55f9ffe1f36d102dbf060aa0cca3fb816a8bc9d48d5b"),
		createFileMultiHash("go", 12688957, "fc4efb561ebedfb03e607c141c0bdc53", "80267009547763ead36a18080b5f39cf18f7810a", "64e4bf4b8091fe126aff1841019a7f30815742cc4933faff6279475246fc27df", "94d94fa7b3b0c94371bc2e43186dd0d12c588f8a3a69f7f6479cb7bad6a2cf47983b118daa5bfa6cbf49a9778249ed9f3154f0d8f9e8f51d2d75deaf4427bbd0"),
		createFileMultiHash("wget", 585416, "a89b3333b14719b8761760541cf73849", "c4a2729b8d3d16e20f2f63f4048e4918150c5990", "57f843898368aef4f112e76ab96271360de9a412a41ea2a238535f42ccb8bc99", "f93f5fe6d481463f5708c85190677710dd4ba036f84d86ac37030e1bc0543527a4ab87a25d15611943b4cdc507b81c537a345148a9719947cca2bb2992aab258"),
		createFileMultiHash("nano", 258896, "bf7cdf52a8af2e62239aa49801660213", "62d17c368ccd6902a8b65bfb48f0cb6aaf0a2148", "67845d0332148d310c2ab4050a411d09a145598dba3753e6defda7d1b752b43b", "84ef11f8290e374cb12db4c1a13950eb0c344f3a1ec111061bfad475f12986b338ce8fb4aa7e26744a81ddd0cc0dcbfeddd1c9be6f43182f89c20f268cdc14a5"),
		createFileMultiHash("zypper", 3411072, "2208079a0901eaf5341c2a3814d20720", "644d68f6a72755aeb94479bfa8c04e3714bd9006", "0238275545613034fb30154cfc0b5188bd2f28dbe7dddb0ad57ee498e55083f6", "d79df6e57d02afc2a73356c6b2863a4f37689fbe35266c0bec237aabce015a23946ad67adf96bc5125e85cca41985ac75d4f7845b22e073eafba79cfc7264ffa"),
	}
	imgs = []*core.FileMultiHash{
		createFileMultiHash("104481404_2269289369892972_6614553252192578924.jpg", 418353, "6de8a2227cb2bcfecfa87538f9fd8663", "7b6be8974cd67afcfbe660528c85c12fa5b1ba01", "3eddef97785aaa6b947b2b35bc765a2c8924300e40f4118781d1f60aff4ef3ea", "aedbe4b29a028509ad2028eaede6e8616fecb28ddd30de97f3f17c85a1bf8dfa8b96383c9e923e4becf67a314cb1f097275b1d539339c71fb25f6e961473b23a"),
		createFileMultiHash("120865501_4306510232713090_6524281912074268580.jpg", 158742, "5ebbd09fc8c00a0bfe2da8d7e5045c1c", "fe106a4a499e7e61324eed072659f99c070a0d74", "76e78c5113b34084b619c0c6f4c7d3d6d403bced1673ed3696d1a3374404e2f8", "c75f63e2e34cf83babba313c485a04bb89e4ecb18a5e8d6e84093da6f268078f1d45db8c88e9b4be2bb548db4e4a3477cb6d101d3c0ff4d48269d20f1498711c"),
		createFileMultiHash("122679148_2779317922314746_8369294723265122474.jpg", 509596, "c8898a03d971f7b4c0012a6dcae25b76", "e7905db63d4e79188a01e74def34bbbc43c06d7d", "48da9d0ecbf7a3618e3e2ed7543ab8a7204ab3538afd2a3750adccec8546b76f", "d8bd70cc8a887210366b99df19c69ae6a98349da5b99fc69b23eed424938d9d784029f0e3f0cb0d18c5b1716b0cd8a970dd85551c22dc8b782aa36d3a90d0f2e"),
		createFileMultiHash("281827975_3848323758726757_4936926896804375945.jpg", 715758, "b6c6638e931300378548de28b6e47ff7", "7f0f3d5ea0d101c09ebd7cee89e557f93fb809e4", "112f2418f6a2a368900bcd00b7577776126f6ed2fe7fd548b842c2c040c817b8", "8b48901b290091bd50ef68575d13f0384758f6edf38098b33917dcc74a10a19dea8c30b4e772eec0d68caa77e022addaf0c746f4fef0e8a3b4d45c997923c0db"),
		createFileMultiHash("284589156_10228821387030040_5461559314640999219.jpg", 167500, "729f843c44c27cdf9a52bf74292458fa", "c009fb217dead43ffd26f7ecf722a4775c5eb23a", "e83e1af6a843b156e8262468a44b5d81daf74233a407c3b0c2a63e5c3a4ed0e4", "2fc906222e28b498e5c4ab14c1cc579ec51dde536ddc73e3583caae4656116088962f94cb161d4985b989eaac3985255cd4a1a343d696fc65e18aad2c92f128a"),
		createFileMultiHash("347395508_813195557036681_967528494166602120.jpg", 78053, "5c2552b0eb114751bfdaedc21ca3702d", "f045f68c06291469662dda574c3609163904632c", "4bd49dce7fd37445167920bc8195aa1a467b1ac7d00b4d6415d5e3d5a155a33c", "95bcde9f6d4c7c1914aa04b85d39008c5f2fd82d26452dfe3cee9ee16a8f142dad041be3266b30baf3369a189432ca076fd88cc00df0d3f8bd8e708e24d30dc4"),
		createFileMultiHash("347431900_816433553377073_6154863431794469435.jpg", 231260, "bb79303a4fbd90b21a2be19f68d417dc", "3c4779633c2b23da2cb163ed70efbaa38f3e2ce9", "1890d7fe616b1d3f0a32a27d13fc02220c5d60c6d26d18765c4dd3847b1b1232", "8b584bba277bfcdfc203bbbbd7777f7f75e0f34db9573b2d1721a1e9e46c605af7988a6d30a4ceca5d0d4c214afef4337f89e78b7f9252e468e385842bfd1062"),
		createFileMultiHash("349104259_958682395315943_6289452280342549092.jpg", 406806, "88c574b48ce09aa06efcd77e10a42e14", "6daf18ee99201ede31ffcf0608dd4f155a31d85d", "2d93133f7ee7b3a60a04b9e76c33af37a4fcd922321dadf949cd4c0010164fa3", "97fba7735c6bbc7674c1ab7deecf62db3a9ce9e98edd93167b0757866313ec881da56523a9750eff73718afbf258102a523878a7f2a1ccdf718f1993ed09683f"),
		createFileMultiHash("383446133_871775087647419_7685832678178040742.jpg", 191024, "a22d6e54b78104c42e4b38d417c0632c", "db29268cf736775a843858f09cb7b10c9af2df81", "95ddb3ef9a28922bd8145f33acefd54457f506327a0264c46e15c5072dd85cc6", "1b156ec93d0cf0e9e1389e7cfe5d99ec324afb2e066661d544b69104399bc5f1362aa3e8f7eb2801448809306c4f3e70dd38bb8730e3af30c7ca1d542bb2da67"),
		createFileMultiHash("417404879_943973973956363_9043642269447999040.jpg", 248458, "a87b3233252bb09825099cce022a430a", "119f3e1f111d6260ecfb9e65c8b3055d0c6da5c0", "edbd91203f1196eb8a3cd014df1d22bfa76ea15042a0d29a9d805c52d2af076c", "a153da6ba56a3e9162b39e8801ead8c173f9fd14322b62dd9112d9d8f99d93799dda36687242de6af4624081d150c9edc9bd3ae25e9f635e4c2cbfa9b8e417ca"),
		createFileMultiHash("431000064_815846923916849_548192370849311731.jpg", 129426, "a3a1219018cf2593a28c4434e3a7ee36", "1032368bcad2210067d3bb907f69aca134759814", "e23b03d550cbf5ee4a4739b66e8a0d78b33af2367a396fe6d4d323083db34996", "b81582de53565a458da054b352be2ffe3c6835ff41d336d30928e1be17658bdd039d0fbd23a789be643561b5edaf3da5feacf7635eff9271b749dcae2d52b094"),
		createFileMultiHash("457251862_902488768586469_4412682217370698292.jpg", 290825, "b79b99144006bab47207fb84d85339c8", "3c6e4f08c3e5f848959a7ef9a1878142d692d0a4", "fdcb69f9cb7c669b4eb8d059ef6da6debaf66aa2ac5225fdc7337ebfe3545f7b", "a9e81bce5281050e50f913f85ca2d278cc04049544bf5d4f2eaa40bb3c3bf2d57efe90eb63fce8b6907c5dd5632074a16d061be3ba88225c6570633fee59ff0e"),
		createFileMultiHash("461989168_946678077487846_1616035097810312250.jpg", 271906, "76f741d75f413f532ee07e59d03c78f4", "4eee64943153a1e1bce2021c62de3ed34deebc75", "e877690ff8aaebcd7f2bbffa996f236d49c1d1a4a17921dbb5381373184e87a6", "fecfbe286c86e400a55ca926ecd87ab1809d1d82e108e62b445a2a8be2c48d0ac3592fa1b6be24a5b9ddba70334545ac83b4fb47fc4d554fef46dbe5337771c8"),
		createFileMultiHash("470667312_918872370344734_2712499210299194559.jpg", 429299, "a45bad4c94d0aae050a161c030cc6811", "da8620bf57cdb3d6a31f3b73a53e47a0ec1c3cfb", "99d4a0eb96f8e32108afbaa74a4b87c6b6e25da140b8c306a2e62177021f27b3", "0f1b19c48a9654eddd96143a9329f48740ff3f01ec032b7bebd3c600e3557c15c86a3d08b765077a7143be4c9f32dfaa805e90221f78474b683bbee46f1bdc0b"),
		createFileMultiHash("471139620_1137339247755450_8509052910057230594.jpg", 538691, "6211df7f5fcacfdb14840a2cfbeab080", "a9a3f17d9a20e8047219655530faff1b2e0ec9f5", "dba697716886e4326209cb44694d25106f8a3e258e9a06ae2b1939a0eff877b6", "e5edc51326b04859671891d374fee628e4783ad7d2d85acd51f9beb0ea9641c43667058243057a75ee658a0843bafe45e045b7b1736b64a26076e2cb671d6357"),
		createFileMultiHash("471558929_10162384100542454_2069148967528258482.jpg", 436148, "dbf8d8d07f1f419321f3f9a20c6fbde0", "291ce9b1ddbf84bfa8a245e26247b582e4100de5", "567be51a391283dad022d2ad89fd1f67ec1b63652cb240761ad9a3e9e57a9f32", "b3bb43c12238f4934495cb3cca1b0ed7c0fb9b8762b1914907a5c96bd29d40f8f0abe2169121440eb920e14a87e10efaf4dc60c3fbd1dc02734ba4867dd13010"),
		createFileMultiHash("473314735_1116165696638735_6786896280360675715.jpg", 493774, "badb77e95a266d77153ca92ce74b000b", "1f284867924e1c1aed8f92369268c95b79aaf434", "8098739e5d6c78d6d2f1e8ea40065b816875d0918dd0cf1690afbe181cff6bb2", "739b7536f131f0fbb9e186a241d2bd2182eb796a9eeeced43d0f70b98721ba495cf97daaf30bdeab889b09a8042c47f8454eab728e04f0778c4dbc23ab46ac60"),
		createFileMultiHash("473314735_1116165696638735_6786896280360675715_copy.jpg", 493774, "badb77e95a266d77153ca92ce74b000b", "1f284867924e1c1aed8f92369268c95b79aaf434", "8098739e5d6c78d6d2f1e8ea40065b816875d0918dd0cf1690afbe181cff6bb2", "739b7536f131f0fbb9e186a241d2bd2182eb796a9eeeced43d0f70b98721ba495cf97daaf30bdeab889b09a8042c47f8454eab728e04f0778c4dbc23ab46ac60"),
	}
	return
}
