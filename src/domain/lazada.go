package domain

import (
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"regexp"
)

type LazadaForwarder struct {
	Caller Caller
	Logger Logger
}

func NewLazadaForwarder(caller Caller, logger Logger) LazadaForwarder{
	return LazadaForwarder{Caller: caller, Logger: logger}
}

func (h LazadaForwarder) GetShopId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shopUrl := vars["shop_url"]
	if match, _ := regexp.MatchString("https://(www\\.|)lazada.vn/shop/.*", shopUrl); !match {
		h.Caller.SendError(w, http.StatusBadRequest, "invalid shop url")
		return
	}
	res, err := h.Caller.Get(shopUrl, nil, map[string]interface{}{
		"User-Agent" : "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:86.0) Gecko/20100101 Firefox/86.0",
		"Cookie": "lzd_cid=192df3dc-b096-4c3b-e0b3-b5659ed1090b; t_uid=192df3dc-b096-4c3b-e0b3-b5659ed1090b; hng=VN|en|VND|704; userLanguageML=en; lzd_sid=147fbd576c6c16c68258c67a2a9c89ac; _m_h5_tk=f637454a047e6cb43e4bcfab5595693d_1615546841316; _m_h5_tk_enc=4f577956ea2989cb7b440713ff35b4b1; _tb_token_=eb75ee565155a; _bl_uid=UhkUmmwv6p611ma4syXqf4ges1d8; t_fv=1615537481251; t_sid=pRoyP10k04kQ1Vg9LKGV2ORsj36D2NmY; utm_origin=https://www.google.com/; utm_channel=SEO; cna=SRPSGGoRmwICATq6Mrg14urQ; _gcl_aw=GCL.1615537482.EAIaIQob…-X1ByxhS3Ye7WHqDGZE3lZcdPvVa4iWc5TGf_g55ZFjmvQwWsVjUU_Kos0v6pxC.; isg=BLu7TZs4X4mecWM9uOOlZzn4SZYlEM8SN-mFw614l7rRDNvuNeBfYtlMIjRCNycK; l=eBryxWcVjR8qG8lzBOfahurza77OSIOYYuPzaNbMiOCPOYCH5UzRW6NgvGTMC36Nh6oMR350TVMyBeYBYQOSnxvtOKLUVuMmn; xlly_s=1; EGG_SESS=S_Gs1wHo9OvRHCMp98md7FYa2w9hcDjT9-x6pE3w64L_CExkbAVqU62tjZvgGRlYSKVWEveRYQyVHF_XIB3jnYggEt35Dpe91NXbKLrk_C7z1mllgG-wCq5o5NYQbOLum-hJ8oRRiZDNKa9GLdtDXvEDOF95bS5CYQygT9OCkwo=; _uetsid=655a72f0830c11eb96eec7106ff80838; _uetvid=655a8770830c11eb81a44b64d285f06c",
	})
	if err != nil {
		h.Caller.SendError(w, res.StatusCode, err.Error())
		return
	}
	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		h.Caller.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	re, _ := regexp.Compile("shopId=([0-9]+)")
	shopId := re.Find(bytes)
	h.Logger.Info("%v", string(shopId))

	w.Write(bytes)
}

