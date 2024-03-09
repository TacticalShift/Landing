
const CommunityURLs = [
	{
		name: "discord",
		url: "https://discord.gg/bqnGUCmAgU"
	},
	{
		name: "teamspeak",
		url: "ts3server://ts3.tacticalshift.ru?port=9987"
	},
	{
		name: "vk",
		url: "https://vk.com/tacticalshift"
	},
	{
		name: "steam",
		url: "https://steamcommunity.com/groups/tacticalshift"
	},
	{
		name: "youtube",
		url: "https://www.youtube.com/channel/UCUCo9u8-z0pLinUEWzSlPmg"
	}
]


function setCommunityURLs() {
	CommunityURLs.forEach((url) => {
		document
			.querySelectorAll(`[meta-url=${url.name}]`)
			.forEach((link) => link.href = url.url)	
	})
};

document.addEventListener("DOMContentLoaded", setCommunityURLs);
