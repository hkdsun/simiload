require 'open3'

class LoadGenerator
  def initialize
    @generators = []
  end

  def add_load(shop_id, concurrency: 4, rps: 10, duration: "5m")
    cmd = "hey -c #{concurrency} -q #{rps} -z #{duration} 'http://127.0.0.1:8080/shop/#{shop_id}'"

    stdin, stdout, stderr, wait_thr = Open3.popen3(cmd)
    stdin.close

    @generators << { stdout: stdout, stderr: stderr, wait_thr: wait_thr, shop_id: shop_id }
  end

  def wait
    puts "Waiting for generators to terminate"

    @generators.each do |gen|
      puts
      puts
      puts
      puts
      puts "************ shop_id=#{gen[:shop_id]} **************"
      puts "****************************************************"
      puts
      gen[:wait_thr].join

      puts gen[:stdout].read
      puts gen[:stderr].read
    end
  end

  def register_signal_handlers
    Signal.trap("INT") do
      @generators.each do |gen|
        gen[:wait_thr].kill
      end
      exit
    end
  end
end

gen = LoadGenerator.new
gen.register_signal_handlers

puts "starting base load on shops 1,2,3,4"

gen.add_load(1)
gen.add_load(2)
gen.add_load(3)
gen.add_load(4)

sleep 30

puts "starting flash sale on shop 5"

gen.add_load(5, concurrency: 40, rps: 2000, duration: "30s")

gen.wait
