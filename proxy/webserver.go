package proxy
import (
	"net"
	"net/http"
	"strings"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
)

func (p *Proxy) startInternalWeb(){
	if p.Config.WebPort >0 {
		go func() {
			p.pairs=make(map[string]string)
		router := gin.Default()
    router.GET("/", p.getAll)
    router.GET("/add", p.add)
	router.GET("/del", p.del)
    router.Run("localhost:"+strconv.Itoa(int(p.Config.WebPort)))
	}()
}
}

func (p *Proxy) add(c *gin.Context) {
	domain := c.Query("d")
	ip := c.Query("ip")
	if domain != "" && ip != "" {
		p.pairs[domain] = ip
	}
}
func (p *Proxy) del(c *gin.Context) {
	domain := c.Query("d")
	if domain != "" {
		delete(p.pairs, domain)
	}
}
func (p *Proxy) getAll(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, p.pairs)
}
func (p *Proxy) lookup(domain string) net.IP {
	d_without_dot := strings.TrimRight(domain, ".")
	if val, ok := p.pairs[d_without_dot]; ok {
		return net.ParseIP(val)
	}
	return nil

}

func (p *Proxy) LookInternal(req *dns.Msg) (resp *dns.Msg) {

	q := req.Question[0]
	qname :=q.Name
	if q.Qtype == dns.TypeA {
		ip := p.lookup(qname)
		if ip != nil {
			p.logger.Info("Cache Hit internal for ",qname,ip)
			a := dns.A{
				Hdr: dns.RR_Header{
					Name:   qname,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				A: ip.To4(),
			}
			m := new(dns.Msg)
			m.SetReply(req)
			m.Authoritative = true
			m.Answer = append(m.Answer, &a)
			return m
		}
		return nil
	}
	return nil
}