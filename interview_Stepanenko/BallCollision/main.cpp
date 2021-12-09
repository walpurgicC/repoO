#include "SFML/Graphics.hpp"
#include "MiddleAverageFilter.h"

constexpr int WINDOW_X = 1024;
constexpr int WINDOW_Y = 768;
constexpr int MAX_BALLS = 300;
constexpr int MIN_BALLS = 100;
const float PI = 3.141592653589793;

Math::MiddleAverageFilter<float,100> fpscounter;

struct Ball
{
    sf::Vector2f p;
    sf::Vector2f dir;
    float r = 0;
    float speed = 0;
};

void draw_ball(sf::RenderWindow& window, const Ball& ball)
{
    sf::CircleShape gball;
    gball.setRadius(ball.r);
    gball.setPosition(ball.p.x, ball.p.y);
    window.draw(gball);
}

void move_ball(Ball& ball, float deltaTime)
{
    float dx = ball.dir.x * ball.speed * deltaTime;
    float dy = ball.dir.y * ball.speed * deltaTime;
    ball.p.x += dx;
    ball.p.y += dy;
}

void draw_fps(sf::RenderWindow& window, float fps)
{
    char c[32];
    snprintf(c, 32, "FPS: %f", fps);
    std::string string(c);
    sf::String str(c);
    window.setTitle(str);
}

bool has_collision(Ball& first, Ball& second)
{
    float r = first.r + second.r;
    r = r * r;
    float dist = (first.p.x - second.p.x) * (first.p.x - second.p.x) + (first.p.y - second.p.y) * (first.p.y - second.p.y);
    return r >= dist;
}

void resolve_collision(Ball& first, Ball& second)
{
    sf::Vector2f dt = first.p - second.p;
    float dist = sqrt(dt.x * dt.x + dt.y * dt.y);
    sf::Vector2f mtd = dt * (((first.r + second.r) - dist) / dist);
    float mtd_len = sqrt(mtd.x * mtd.x + mtd.y * mtd.y);
    sf::Vector2f mtd_normilized = mtd / mtd_len;

    float inverse_mass_first = 1 / (first.r * first.r * PI);
    float inverse_mass_second = 1 / (second.r * second.r * PI);

    first.p += mtd * (inverse_mass_first / (inverse_mass_first + inverse_mass_second));
    second.p -= mtd * (inverse_mass_second / (inverse_mass_first + inverse_mass_second));

    sf::Vector2f velocity_first = first.dir * first.speed;
    sf::Vector2f velocity_second = second.dir * second.speed;

    sf::Vector2f rel_velocity = velocity_first - velocity_second;
    float vn = rel_velocity.x * (mtd_normilized.x) + rel_velocity.y * (mtd_normilized.y);

    if (vn > 0.0f)
        return;

    float i = (-(1.0f + 1) * vn) / (inverse_mass_first + inverse_mass_second);
    sf::Vector2f impulse = mtd_normilized * i;

    velocity_first += impulse * inverse_mass_first;
    float velocity_first_len = sqrt(velocity_first.x * velocity_first.x + velocity_first.y * velocity_first.y);
    velocity_second -= impulse * inverse_mass_second;
    float velocity_second_len = sqrt(velocity_second.x * velocity_second.x + velocity_second.y * velocity_second.y);

    first.dir.x = velocity_first.x / velocity_first_len;
    first.dir.y = velocity_first.y / velocity_first_len;
    first.speed = velocity_first_len;

    second.dir.x = velocity_second.x / velocity_second_len;
    second.dir.y = velocity_second.y / velocity_second_len;
    second.speed = velocity_second_len;
}

int main()
{
    sf::RenderWindow window(sf::VideoMode(WINDOW_X, WINDOW_Y), "ball collision demo");
    srand(time(NULL));

    std::vector<Ball> balls;

    // randomly initialize balls
    for (int i = 0; i < (rand() % (MAX_BALLS - MIN_BALLS) + MIN_BALLS); i++)
    {
        Ball newBall;
        newBall.r = 5 + rand() % 5;
        newBall.p.x = rand() % WINDOW_X;
        newBall.p.y = rand() % WINDOW_Y;
        newBall.dir.x = (-5 + (rand() % 10)) / 3.;
        newBall.dir.y = (-5 + (rand() % 10)) / 3.;
        newBall.speed = 30 + rand() % 30;
        balls.push_back(newBall);
    }

   // window.setFramerateLimit(60);

    sf::Clock clock;
    float lastime = clock.restart().asSeconds();

    while (window.isOpen())
    {

        sf::Event event;
        while (window.pollEvent(event))
        {
            if (event.type == sf::Event::Closed)
            {
                window.close();
            }
        }

        float current_time = clock.getElapsedTime().asSeconds();
        float deltaTime = current_time - lastime;
        fpscounter.push(1.0f / (current_time - lastime));
        lastime = current_time;

        /// <summary>
        ///  Для улучшения архитектуры кода можно выделить класс Ball и поместить в него все данные и методы, связанные с шариком.
        ///  Для улучшения алгоритма поиска и разрешения коллизий можно попробовать разделить все доступное пространство на сегменты определенного размера
        ///  И уже в этих сегментах искать коллизии.
        float eps = 1e-10;
        for (int i = 0; i < balls.size(); i++)
        {
            if (balls[i].p.x < eps)
            {   
                balls[i].p.x += 1;
                balls[i].dir.x = -balls[i].dir.x;
            }
            else if (balls[i].p.x + 2 * balls[i].r > WINDOW_X - eps)
            {
                balls[i].p.x -= 1;
                balls[i].dir.x = -balls[i].dir.x;
            }

            if (balls[i].p.y < eps)
            {
                balls[i].p.y += 1;
                balls[i].dir.y = -balls[i].dir.y;
            }
            else if (balls[i].p.y + 2 * balls[i].r > WINDOW_Y - eps)
            {
                balls[i].p.y -= 1;
                balls[i].dir.y = -balls[i].dir.y;
            }
        }
        
        for (int i = 0; i < balls.size(); i++)
        {
            for (int j = i + 1; j < balls.size(); j++)
            {
                if (has_collision(balls[i], balls[j]))
                {
                    resolve_collision(balls[i], balls[j]);
                }
            }
        }

        for (auto& ball : balls)
        {
            move_ball(ball, deltaTime);
        }

        window.clear();
        for (const auto ball : balls)
        {
            draw_ball(window, ball);
        }
        sf::Vector2f pos(0, 0);
        sf::Vector2f direc(0, 0);

        Ball ball;
        ball.p = pos;
        ball.dir = direc;
        ball.r = 1;
        ball.speed = 0;
        draw_ball(window, ball);
		//draw_fps(window, fpscounter.getAverage());
		window.display();
    }
    return 0;
}